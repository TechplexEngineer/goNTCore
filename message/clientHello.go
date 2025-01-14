// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"

	"github.com/techplexengineer/gontcore/entryType"
)

//ClientHello is sent when a client is first communicating to a server.
type ClientHello struct {
	message
	protocol   [2]byte
	clientName *entryType.String
}

//implements the stringer interface
func (m ClientHello) String() string {
	return fmt.Sprintf("Client Hello - protocol:%#x name:%s", m.protocol, m.ClientName())
}

func (m ClientHello) ClientName() string {
	return m.clientName.String()
}

func (m ClientHello) Protocol() [2]byte {
	return m.protocol
}

//NewClientHello creates a new instance of ClientHello with the specified Protocol and Client Name.
func NewClientHello(protocol [2]byte, clientName string) *ClientHello {
	return &ClientHello{
		message: message{
			mType: MTypeClientHello,
		},
		protocol:   protocol,
		clientName: entryType.NewString(clientName),
	}
}

//GetProtocol gets the protocol of the message.
func (ch *ClientHello) GetProtocol() [2]byte {
	return ch.protocol
}

//GetClientName gets the name of the client from the message.
func (ch *ClientHello) GetClientName() string {
	return ch.clientName.GetValue()
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (ch *ClientHello) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{ch.Type().Byte()})
	if err != nil {
		return err
	}
	_, err = writer.Write(ch.protocol[:])
	if err != nil {
		return err
	}
	err = ch.clientName.MarshalEntry(writer)
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (ch *ClientHello) UnmarshalMessage(reader io.Reader) error {
	ch.mType = MTypeClientHello
	protoBuf := make([]byte, 2)
	st := new(entryType.String)

	_, err := io.ReadFull(reader, protoBuf)
	if err != nil {
		return err
	}
	err = st.UnmarshalEntry(reader)
	if err != nil {
		return err
	}

	copy(ch.protocol[:], protoBuf)
	ch.clientName = st
	return nil
}
