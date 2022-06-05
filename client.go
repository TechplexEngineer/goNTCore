// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goNTCore

import (
	"fmt"
	"github.com/techplexengineer/gontcore/entryType"
	"github.com/techplexengineer/gontcore/storage"
	"github.com/techplexengineer/gontcore/util"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/techplexengineer/gontcore/message"
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
	// ClientStartingSync indicates that the client has received the
	// server hello and is beginning to synchronize values with the
	// server.
	ClientStartingSync
	// ClientReady indicates that the client has competed initial sync
	// and is ready to respond to queries.
	ClientReady
)

//type IStorageEntry interface {
//	// Get the name of the table entry
//	GetName() string
//	// Get the unique ID of the table entry
//	GetID() uint16
//	// Get the serial number of the table entry
//	GetSN() uint16
//	// Check if the entry is persistent
//	// If true, Servers sill store the most recent value of this entry and load that entry on next start
//	IsPersistent() bool
//	// The underlying entry in the table
//	GetEntry() entryType.Entrier
//}
//
//type INetworkTable interface {
//	Assign(ea *message.EntryAssign)
//	Update(entryId uint16, entrySN uint16, entry entryType.Entrier)
//	DeleteID(entryId uint16)
//	DeleteKey(key string)
//	DeleteAll()
//	NumEntries() int
//	GetEntry(key string) (IStorageEntry, error)
//	GetKeys() []string
//}

type ListenerCallback func(entry storage.StorageEntry)

//Client is a Network Table client.
type Client struct {
	net.Conn
	connected bool
	name      string
	status    ClientStatus
	storage   *storage.NetworkTable
	Debug     bool

	// we should not modify seedData
	// Optionally passed in to ensure the keys exist
	seedData *storage.NetworkTable

	listeners    map[int]Listener
	listenerLock *sync.RWMutex
}

//NewClient creates a new client to communicate to a Network Table server.
// serverHost should be an IP address or hostname eg: localhost, 127.0.0.1, roborio-4909-frc.local
// name is used to identify this client with the networktables server. Generally want to keep it short, most implementations ignore the value
// initialData is optional. One can pass a table with existing data the client wants to ensure the server has
func NewClient(serverHost string, name string, initialData *storage.NetworkTable) (*Client, chan error, error) {
	//log.Printf("Connecting to %s", net.JoinHostPort(serverHost, strconv.Itoa(PORT)))
	conn, err := net.Dial("tcp", net.JoinHostPort(serverHost, strconv.Itoa(PORT)))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to start tcp connection - %w", err)
	}
	c := &Client{
		Conn:         conn,
		connected:    true,
		name:         name,
		status:       ClientDisconnected,
		storage:      storage.NewNetworkTable(),
		seedData:     initialData,
		listeners:    map[int]Listener{},
		listenerLock: new(sync.RWMutex),
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
	if c.Debug {
		log.Printf("<=== Sent %s", msg)
	}
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
	if c.Debug {
		log.Printf("===> Got %s - %s", msg.Type().String(), msg)
	}
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
		c.storage.Assign(m)

		c.processListeners(m.GetName())
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
		c.storage.Update(m.EntryID(), m.EntrySN(), m.Entry())

		if name, err := c.storage.IdToName(m.EntryID()); err == nil {
			c.processListeners(name)
		}

	case message.MTypeEntryFlagUpdate:
		log.Print("Entry Flag Update Not Implemented")
	case message.MTypeEntryDelete:
		m := msg.(*message.EntryDelete)
		c.storage.DeleteID(m.EntryID())
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

// this should be called after any entry is added or changed
func (c *Client) processListeners(changedKey string) {
	changedKey = util.SanitizeKey(changedKey)
	c.listenerLock.RLock()
	defer c.listenerLock.RUnlock()
	for handle, handler := range c.listeners {
		if handler.prefix && strings.HasPrefix(changedKey, handler.key) || handler.key == changedKey {
			entry, err := c.storage.GetEntry(changedKey)
			if err != nil {
				log.Printf("Key %s missing when preparing listener callback handle: %d", changedKey, handle)
				continue
			}
			handler.callback(entry)
		}
	}

}

//find the differences and send EntryAssign message to the server for each
func (c *Client) notifyOfDifference() {
	if c.seedData != nil && c.seedData.NumEntries() == 0 {
		if c.Debug {
			log.Printf(">><< Nothing to send")
		}
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
	_, err := c.storage.GetEntry(key)
	return err == nil
}

// Remove entry that exactly matches key
func (c *Client) Delete(key string) {
	c.storage.DeleteKey(key)
}

// Remove all entries from the table
func (c *Client) DeleteAll() {
	c.storage.DeleteAll()
}

// true if the value is persisted, false otherwise
// it the value is missing, false is returned
func (c *Client) IsPersisted(key string) bool {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return false
	}
	return entry.IsPersistent()

}

// Get a list of all the keys in the table
func (c *Client) GetKeys() []string {
	return c.storage.GetKeys()
}

// Returns the boolean the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetBoolean(key string, def bool) bool {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeBoolean {
		return def
	}
	e := entry.GetEntry().(*entryType.Boolean)
	return e.GetValue()
}

// Put a boolean in the table.
// Returns: False if the table key already exists with a different type
func (c *Client) PutBoolean(key string, val bool) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewBoolean(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.Boolean)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the number the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetNumber(key string, def float64) float64 {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeDouble {
		return def
	}
	e := entry.GetEntry().(*entryType.Double)
	return e.GetValue()
}

// Put a number in the table.
func (c *Client) PutNumber(key string, val float64) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewDouble(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.Double)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the string the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetString(key string, def string) string {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeString {
		return def
	}
	e := entry.GetEntry().(*entryType.String)
	return e.GetValue()
}

// Put a string in the table.
func (c *Client) PutString(key string, val string) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewString(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.String)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the raw value (byte array) the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetRaw(key string, def []byte) []byte {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeRawData {
		return def
	}
	e := entry.GetEntry().(*entryType.RawData)
	return e.GetValue()
}

// Put a raw value (byte array) in the table.
func (c *Client) PutRaw(key string, val []byte) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewRawData(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.RawData)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the boolean array the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetBooleanArray(key string, def []bool) []bool {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeBooleanArray {
		return def
	}
	e := entry.GetEntry().(*entryType.BooleanArray)
	return e.GetValue()
}

// Put a boolean array in the table.
func (c *Client) PutBooleanArray(key string, val []bool) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewBooleanArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.BooleanArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the number array the key maps to.
func (c *Client) GetNumberArray(key string, def []float64) []float64 {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeDoubleArray {
		return def
	}
	e := entry.GetEntry().(*entryType.DoubleArray)
	return e.GetValue()
}

// Put a number array in the table.
func (c *Client) PutNumberArray(key string, val []float64) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewDoubleArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.DoubleArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

// Returns the string array the key maps to. If the key does not exist or is of different type, it will return the default value.
func (c *Client) GetStringArray(key string, def []string) []string {
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		return def
	}
	if entry.GetEntry().Type() != entryType.ETypeStringArray {
		return def
	}
	e := entry.GetEntry().(*entryType.StringArray)
	return e.GetValue()
}

// PutStringArray Puts a string array in the table.
func (c *Client) PutStringArray(key string, val []string) bool {
	// check if the key exists, if so send update else send assign
	entry, err := c.storage.GetEntry(key)
	if err != nil {
		// entry does not exist
		newEntry := entryType.NewStringArray(val)
		msg := message.NewEntryAssign(key, newEntry, false, message.GetNewEntryID(), message.GetNewEntrySN())
		c.SendMsg(msg)
	} else {
		// entry exists
		e, ok := entry.GetEntry().(*entryType.StringArray)
		if !ok {
			return false
		}
		e.SetValue(val)
		msg := message.NewEntryUpdate(entry.GetID(), entry.GetSN()+1, e)
		c.SendMsg(msg)
		c.storage.Update(entry.GetID(), entry.GetSN()+1, e)
	}
	return true
}

func (c *Client) GetSnapshot() []storage.SnapShotEntry {
	return c.storage.GetSnapshot()
}

type Listener struct {
	// key to listen to
	key string
	// callback to call
	callback ListenerCallback
	// do prefix matching
	prefix bool
}

// AddPrefixListener creates a Listener for changes to any keys that begin with prefix
// returns an integer handle which can be used to remove the listener
func (c *Client) AddPrefixListener(prefix string, callback ListenerCallback) int {
	return c.addListener(prefix, callback, true)
}

// AddKeyListener Creates a Listener for changes to a specific key
// returns an integer handle which can be used to remove the listener
// only receives calls when the server changes the value. Local changes do not trigger the callback
func (c *Client) AddKeyListener(key string, callback ListenerCallback) int {
	return c.addListener(key, callback, false)
}

func (c *Client) addListener(key string, callback ListenerCallback, prefixMatch bool) int {
	key = util.SanitizeKey(key)
	c.listenerLock.Lock() // lock for writing
	defer c.listenerLock.Unlock()

	// find the highest currently used handle
	largestHandle := 0
	for h, _ := range c.listeners {
		if h > largestHandle {
			largestHandle = h
		}
	}
	// increment to get next handle
	nextHandle := largestHandle + 1
	c.listeners[nextHandle] = Listener{
		key:      key,
		callback: callback,
		prefix:   prefixMatch,
	}

	return nextHandle
}

// Remove a listener
func (c *Client) RemoveListener(handle int) {
	c.listenerLock.Lock() // lock for writing
	defer c.listenerLock.Unlock()
	delete(c.listeners, handle)
}

//func (c *Client) ContainsTable(key string) bool {}
//func (c *Client) GetTable(key string) NetworkTabler {}
