// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import "io"

//ServerHelloComplete is sent from the server after it has finished telling the new client what entries the server has.
type ServerHelloComplete struct {
	message
}

//implements the stringer interface
func (ServerHelloComplete) String() string {
	return "Server Hello Complete"
}

//NewServerHelloComplete creates a new instance of ServerHelloComplete.
func NewServerHelloComplete() *ServerHelloComplete {
	return &ServerHelloComplete{
		message{
			mType: MTypeServerHelloComplete,
		},
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (shc *ServerHelloComplete) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{shc.Type().Byte()})
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (shc *ServerHelloComplete) UnmarshalMessage(reader io.Reader) error {
	shc.mType = MTypeServerHelloComplete
	return nil
}
