// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"
)

const (
	flagPersistantMask byte = 0x01
)

//EntryFlagUpdate is used to update the flag of an entry.
type EntryFlagUpdate struct {
	message
	entryID    [2]byte
	persistent bool
}

//implements the stringer interface
func (o EntryFlagUpdate) String() string {
	return fmt.Sprintf("FlagUpdate - id:%#x persist:%t", o.entryID, o.persistent)
}

//NewEntryFlagUpdate creates a new instance of EntryFlagUpdate.
func NewEntryFlagUpdate(id [2]byte, persistent bool) *EntryFlagUpdate {
	return &EntryFlagUpdate{
		message: message{
			mType: MTypeEntryFlagUpdate,
		},
		entryID:    id,
		persistent: persistent,
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (efu *EntryFlagUpdate) MarshalMessage(writer io.Writer) error {
	flags := byte(0x00)
	if efu.persistent {
		flags = flags | flagPersistantMask
	}
	_, err := writer.Write([]byte{efu.Type().Byte()})
	if err != nil {
		return err
	}
	_, err = writer.Write(efu.entryID[:])
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{flags})
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (efu *EntryFlagUpdate) UnmarshalMessage(reader io.Reader) error {
	efu.mType = MTypeEntryFlagUpdate
	idBuf := make([]byte, 2)
	flagBuf := make([]byte, 1)

	_, err := io.ReadFull(reader, idBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, flagBuf)
	if err != nil {
		return err
	}

	copy(efu.entryID[:], idBuf)
	efu.persistent = flagBuf[0]&flagPersistantMask == flagPersistantMask
	return nil
}
