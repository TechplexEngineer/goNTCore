// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goNTCore

import (
	"fmt"
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

type DataMap map[string]CacheEntry

type CacheEntry struct {
	entryName       string
	entryID         uint16 //[2]byte
	entrySN         uint16 //[2]byte
	entryPersistent bool
	entry           entryType.Entrier
}

func NewCacheEntryFromEntryAssign(ea *message.EntryAssign) CacheEntry {
	return CacheEntry{
		entryName:       ea.GetName(),
		entryID:         ea.EntryID(),
		entrySN:         ea.EntrySN(),
		entryPersistent: ea.EntryPersistent(),
		entry:           ea.Entry(),
	}
}

func UpdateCacheWithEntryUpdate(entry CacheEntry, m *message.EntryUpdate) CacheEntry {
	entry.entrySN = m.EntrySN()
	entry.entry = m.Entry()
	return entry
}

//Client is a Network Table client.
type Client struct {
	net.Conn
	connected bool
	name      string
	status    ClientStatus
	data      DataMap

	// we should not modify seedData
	// Optionally passed in to ensure the keys exist
	seedData DataMap
}

//NewClient creates a new client to communicate to a Network Table server.
func NewClient(serverHost string, name string, data DataMap) (*Client, chan error, error) {
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
		data:      make(DataMap),
		seedData:  copyDataMap(data),
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

//SendMsg to the connected server
func (c *Client) SendMsg(msg message.Messager) error {
	log.Printf("<=== Sent %s", msg)
	return SendMsg(msg, c)
}

func copyDataMap(origMap DataMap) DataMap {
	newMap := make(DataMap)

	for k, v := range origMap {
		newMap[k] = v
	}

	return newMap
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
		c.data[m.GetName()] = NewCacheEntryFromEntryAssign(m)
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

		// @todo Should we optimize for writes or reads... currently optimized for reads.
		for k, v := range c.data {
			if v.entryID == m.EntryID() {
				if v.entry.Type() != m.Entry().Type() {
					log.Printf("Types differ. Ignoring update")
					break
				}
				log.Printf("Update from:%v to:%v", v.entry, m.Entry())

				c.data[k] = UpdateCacheWithEntryUpdate(c.data[k], m)
				log.Printf("Update to:%v", c.data[k].entry)

				break
			}
		}
	case message.MTypeEntryFlagUpdate:
		log.Print("Entry Flag Update Not Implemented")
	case message.MTypeEntryDelete:
		m := msg.(*message.EntryDelete)
		for k, v := range c.data {
			if v.entryID == m.EntryID() {
				delete(c.data, k)
			}
		}
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
	if len(c.seedData) == 0 {
		log.Printf(">><< Nothing to send")
	} else {
		// @todo
		//Compare c.seedData to c.data send any missing entries
	}
}
