// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

//implements the stringer interface
func (o EntryDelete) String() string {
	return fmt.Sprintf("Delete - ID:%#x", o.entryID)
}

//EntryDelete is used to delete an entry from the network.
type EntryDelete struct {
	message
	entryID [2]byte
}

func (o EntryDelete) EntryID() uint16 {
	return binary.LittleEndian.Uint16(o.entryID[:])
}

//NewEntryDelete creates a new instance of EntryDelete.
func NewEntryDelete(id [2]byte) *EntryDelete {
	return &EntryDelete{
		message: message{
			mType: MTypeEntryDelete,
		},
		entryID: id,
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (ed *EntryDelete) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{ed.Type().Byte()})
	if err != nil {
		return err
	}
	_, err = writer.Write(ed.entryID[:])
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (ed *EntryDelete) UnmarshalMessage(reader io.Reader) error {
	ed.mType = MTypeEntryDelete
	idBuf := make([]byte, 2)

	_, err := io.ReadFull(reader, idBuf)
	if err != nil {
		return err
	}

	copy(ed.entryID[:], idBuf)
	return nil
}
