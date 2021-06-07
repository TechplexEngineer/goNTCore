// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"
)

type MessageType byte

func (o MessageType) String() string {
	switch o {
	case MTypeKeepAlive:
		return "KeepAlive"
	case MTypeClientHello:
		return "ClientHello"
	case MTypeProtoUnsupported:
		return "ProtoUnsupported"
	case MTypeServerHelloComplete:
		return "ServerHelloComplete"
	case MTypeServerHello:
		return "ServerHello"
	case MTypeClientHelloComplete:
		return "ClientHelloComplete"
	case MTypeEntryAssign:
		return "EntryAssign"
	case MTypeEntryUpdate:
		return "EntryUpdate"
	case MTypeEntryFlagUpdate:
		return "EntryFlagUpdate"
	case MTypeEntryDelete:
		return "EntryDelete"
	case MTypeClearAllEntries:
		return "ClearAllEntries"
	case MTypeRPCExecute:
		return "RPCExecute"
	case MTypeRPCResponse:
		return "RPCResponse"
	default:
		return "Unknown Message Type"
	}
}

func (o MessageType) Byte() byte {
	return byte(o)
}

const (
	MTypeKeepAlive           MessageType = 0x00
	MTypeClientHello         MessageType = 0x01
	MTypeProtoUnsupported    MessageType = 0x02
	MTypeServerHelloComplete MessageType = 0x03
	MTypeServerHello         MessageType = 0x04
	MTypeClientHelloComplete MessageType = 0x05
	MTypeEntryAssign         MessageType = 0x10
	MTypeEntryUpdate         MessageType = 0x11
	MTypeEntryFlagUpdate     MessageType = 0x12
	MTypeEntryDelete         MessageType = 0x13
	MTypeClearAllEntries     MessageType = 0x14
	MTypeRPCExecute          MessageType = 0x20
	MTypeRPCResponse         MessageType = 0x21
)

type message struct {
	mType MessageType
}

func (m message) Type() MessageType {
	return m.mType
}

//Marshaler is the interface implemented by types that can marshal themselves into valid Network Table Message.
type Marshaler interface {
	MarshalMessage(io.Writer) error
}

//Unmarshaler is the interface implemented by types that can unmarshal a Network Table Message of themselves.
type Unmarshaler interface {
	UnmarshalMessage(io.Reader) error
}

//Messager is the interface implemented by types that can communicate on the network.
type Messager interface {
	Type() MessageType
	Marshaler
	Unmarshaler
	String() string
}

//Unmarshal takes the type passed in and tries to unmarshal the next bytes from reader based on the type.
//It returns an instance messager.
func Unmarshal(t byte, reader io.Reader) (Messager, error) {
	var msg Messager
	switch MessageType(t) {
	case MTypeKeepAlive:
		msg = new(KeepAlive)
	case MTypeClientHello:
		msg = new(ClientHello)
	case MTypeProtoUnsupported:
		msg = new(ProtoUnsupported)
	case MTypeServerHelloComplete:
		msg = new(ServerHelloComplete)
	case MTypeServerHello:
		msg = new(ServerHello)
	case MTypeClientHelloComplete:
		msg = new(ClientHelloComplete)
	case MTypeEntryAssign:
		msg = new(EntryAssign)
	case MTypeEntryUpdate:
		msg = new(EntryUpdate)
	case MTypeEntryFlagUpdate:
		msg = new(EntryFlagUpdate)
	case MTypeEntryDelete:
		msg = new(EntryDelete)
	case MTypeClearAllEntries:
		msg = new(ClearAllEntries)
	default:
		return nil, fmt.Errorf("unmarshal message: Could not find appropropriate type")
	}
	err := msg.UnmarshalMessage(reader)
	return msg, err
}
