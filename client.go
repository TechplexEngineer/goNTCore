// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goNTCore

import (
	"fmt"
	"github.com/technomancers/goNTCore/cache"
	"github.com/technomancers/goNTCore/entryType"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/technomancers/goNTCore/message"
)

// ClientStatus is the enum type to represent the different
// states/statuses the client could have
type ClientStatus int

const (
	// ClientDisconnected indicates that the client cannot reach
	// the server
	ClientDisconnected ClientStatus = iota
	// ClientStarting indicates that the client has connected to
	// the server but has not began actual communication
	ClientConnected
	// ClientStarting indicates the initial sync is underway
	ClientStarting
	// ClientStartingSync indicates that the client has received the
	// server hello and is beginning to synchronize values with the
	// server.
	ClientStartingSync
	// ClientReady indicates that the client has competed initial sync
	// and is ready to respond to queries.
	ClientReady
)

//Client is a Network Table client.
type Client struct {
	net.Conn
	connected bool
	name      string
	status    ClientStatus
	cache     *cache.NetworkTable

	// we should not modify seedData
	// Optionally passed in to ensure the keys exist
	seedData *cache.NetworkTable
}

//NewClient creates a new client to communicate to a Network Table server.
// serverHost should be an IP address or hostname eg: localhost, 127.0.0.1, roborio-4909-frc.local
// name is used to identify this client with the networktables server. Generally want to keep it short, most implementations ignore the value
// data is optional. One can pass a table with existing data the client wants to ensure the server has
func NewClient(serverHost string, name string, data *cache.NetworkTable) (*Client, chan error, error) {
	//log.Printf("Connecting to %s", net.JoinHostPort(serverHost, strconv.Itoa(PORT)))
	conn, err := net.Dial("tcp", net.JoinHostPort(serverHost, strconv.Itoa(PORT)))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to start tcp connection - %w", err)
	}
	c := &Client{
		Conn:      conn,
		connected: true,
		name:      name,
		status:    ClientDisconnected,
		cache:     cache.NewNetworkTable(),
		seedData:  data,
	}
	echan := make(chan error, 1)
	go c.readIncoming(echan)
	err = c.startHandshake()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to start handshake - %w", err)
	}

	return c, echan, nil
}

//Close closes the connection to the Network Table server.
func (c *Client) Close() error {
	c.connected = false
	c.status = ClientDisconnected
	return c.Conn.Close()
}

// SendMsg to the connected server
// Users will generally want to use the higher level accessor methods Get* and Put*
func (c *Client) SendMsg(msg message.Messager) error {
	log.Printf("<=== Sent %s", msg)
	return SendMsg(msg, c)
}

//Listen for messages sent from the server.
//Best to start this in a go routine.
func (c *Client) readIncoming(echan chan error) {
	for c.connected {
		possibleMsgType := make([]byte, 1)
		_, err := io.ReadFull(c, possibleMsgType)
		if err != nil {
			echan <- fmt.Errorf("unable to read from server - %s", err)
			c.Close()
			continue
		}
		msg, err := message.Unmarshal(possibleMsgType[0], c)
		if err != nil {
			echan <- fmt.Errorf("unable to unmarshal message type %#x - %s", possibleMsgType, err)
			c.Close()
			continue
		}
		c.handler(msg)
	}
}

//StartHandshake starts the handshake with the server.
func (c *Client) startHandshake() error {
	// Step 1: Client sends Client Hello
	return c.SendMsg(message.NewClientHello(ProtocolVersion(), c.name))
}

func (c *Client) handler(msg message.Messager) {
	log.Printf("===> Got %s - %s", msg.Type().String(), msg)
	switch msg.Type() {
	case message.MTypeKeepAlive:
		// keep alive messages keep the underlying TCP connection open and can be ignored
		return
	case message.MTypeServerHello:
		// Step 2: Server replies to ClientHello with ServerHello
		// hello complete, starting sync
		c.status = ClientConnected
		// nothing to send back, server will start sending EntryAssign messages automatically
	case message.MTypeEntryAssign:
		// Step 3: Server sends EntryAssign messages for each entry
		// if we just sent a client hello then the sync is beginning
		if c.status == ClientDisconnected {
			c.status = ClientStartingSync
		}
		m := msg.(*message.EntryAssign)
		c.cache.Assign(m)
	case message.MTypeServerHelloComplete:
		// Step 4: The Server sends a Server Hello Complete message.
		// Server is done sending entryAssigns

		// Step 5: For all Entries the Client recognizes that the Server
		// did not identify with a Entry Assignment.
		// we can now send any entries the server should have
		c.notifyOfDifference()

		// Step 6: The Client sends a Client Hello Complete message.
		err := c.SendMsg(message.NewClientHelloComplete())
		if err != nil {
			log.Printf("unable to send client hello complete - %s", err)
		}
		c.status = ClientReady

	case message.MTypeEntryUpdate:
		m := msg.(*message.EntryUpdate)
		c.cache.Update(m.EntryID(), m.EntrySN(), m.Entry())
	case message.MTypeEntryFlagUpdate:
		log.Print("Entry Flag Update Not Implemented")
	case message.MTypeEntryDelete:
		m := msg.(*message.EntryDelete)
		c.cache.DeleteID(m.EntryID())
	case message.MTypeClearAllEntries:
		log.Print("Clear All Entries Not Implemented")
	case message.MTypeProtoUnsupported:
		m := msg.(*message.ProtoUnsupported)
		log.Printf("Unsupported Protocol Version, server supports %#x", m.GetServerSupportedProtocolVersion())
		c.Close()
	default:
		log.Printf("Unknown message type (%#x) from server", msg.Type())
		c.Close()
	}
}

//find the differences and send EntryAssign message to the server for each
func (c *Client) notifyOfDifference() {
	if c.seedData != nil && c.seedData.NumEntries() == 0 {
		log.Printf(">><< Nothing to send")
	} else {
		// @todo
		//Compare c.seedData to c.data send any missing entries
	}
}

//-----------------------------------------------------------------------------
// User facing accessor methods
//-----------------------------------------------------------------------------

// Check if the table contains a key
func (c *Client) ContainsKey(key string) bool {
	_, err := c.cache.GetEntry(key)
	return err == nil
}

// Remove entry that exactly matches key
func (c *Client) Delete(key string) {
	c.cache.DeleteKey(key)
}

// Remove all entries from the table
func (c *Client) DeleteAll() {
	c.cache.DeleteAll()
}

// true if the value is persisted, false otherwise
// it the value is missing, false is returned
func (c *Client) IsPersisted(key string) bool {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return false
	}
	return entry.EntryPersistent()

}

// Get a list of all the keys in the table
func (c *Client) GetKeys() []string {
	return c.cache.GetKeys()
}

// Returns the boolean the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetBoolean(key string, def bool) bool {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeBoolean {
		return def
	}
	e := entry.Entry().(*entryType.Boolean)
	return e.GetValue()
}

// Put a boolean in the table.
// Returns: False if the table key already exists with a different type
func (c *Client) PutBoolean(key string, val bool) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewBoolean(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.Boolean)
		if !ok {
			return false
		}
		e.SetValue(val)
		log.Printf("C:%d N:%d", entry.EntrySN(), entry.EntrySN()+1)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the number the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetNumber(key string, def float64) float64 {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeDouble {
		return def
	}
	e := entry.Entry().(*entryType.Double)
	return e.GetValue()
}

// Put a number in the table.
func (c *Client) PutNumber(key string, val float64) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewDouble(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.Double)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the string the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetString(key string, def string) string {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeString {
		return def
	}
	e := entry.Entry().(*entryType.String)
	return e.GetValue()
}

// Put a string in the table.
func (c *Client) PutString(key string, val string) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewString(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.String)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the raw value (byte array) the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetRaw(key string, def []byte) []byte {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeRawData {
		return def
	}
	e := entry.Entry().(*entryType.RawData)
	return e.GetValue()
}

// Put a raw value (byte array) in the table.
func (c *Client) PutRaw(key string, val []byte) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewRawData(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.RawData)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the boolean array the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetBooleanArray(key string, def []bool) []bool {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeBooleanArray {
		return def
	}
	e := entry.Entry().(*entryType.BooleanArray)
	return e.GetValue()
}

// Put a boolean array in the table.
func (c *Client) PutBooleanArray(key string, val []bool) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewBooleanArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.BooleanArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the number array the key maps to.
func (c *Client) GetNumberArray(key string, def []float64) []float64 {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeDoubleArray {
		return def
	}
	e := entry.Entry().(*entryType.DoubleArray)
	return e.GetValue()
}

// Put a number array in the table.
func (c *Client) PutNumberArray(key string, val []float64) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewDoubleArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.DoubleArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

// Returns the string array the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetStringArray(key string, def []string) []string {
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.Entry().Type() != entryType.ETypeStringArray {
		return def
	}
	e := entry.Entry().(*entryType.StringArray)
	return e.GetValue()
}

// Put a string array in the table.
func (c *Client) PutStringArray(key string, val []string) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.cache.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewStringArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.Entry().(*entryType.StringArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.EntryID(), entry.EntrySN()+1, e)
		c.SendMsg(msg)
		c.cache.Update(entry.EntryID(), entry.EntrySN()+1, e)
	}
	return true
}

//func (c *Client) ContainsTable(key string) bool {}
//func (c *Client) GetTable(key string) NetworkTabler {}
