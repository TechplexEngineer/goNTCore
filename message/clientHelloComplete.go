// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"io"
)

//ClientHelloComplete is sent to the server after the client finished giving the server all of its entries that does not already exist on the server.
type ClientHelloComplete struct {
	message
}

//NewClientHelloComplete creates a new instance of ClientHelloComplete
func NewClientHelloComplete() *ClientHelloComplete {
	return &ClientHelloComplete{
		message{
			mType: MTypeClientHelloComplete,
		},
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (chc *ClientHelloComplete) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{chc.Type().Byte()})
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (chc *ClientHelloComplete) UnmarshalMessage(reader io.Reader) error {
	chc.mType = MTypeClientHelloComplete
	return nil
}
