// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"

	"github.com/technomancers/goNTCore/entryType"
)

//EntryAssign is used to inform others that a new entry was introduced into the network.
type EntryAssign struct {
	message
	entryName       *entryType.String
	entryID         [2]byte
	entrySN         [2]byte
	entryPersistent bool
	entry           entryType.Entrier
}

// Get a string representation of the EntryAssign object
//implements the stringer interface
func (o EntryAssign) String() string {
	return fmt.Sprintf("%s ID:%#x SN:%#x persist:%t entry:%s", o.entryName, o.entryID, o.entrySN, o.entryPersistent, o.entry)
}

//NewEntryAssign creates a new instance on EntryAssign.
func NewEntryAssign(entryName string, entrier entryType.Entrier, persistent bool, id, sn [2]byte) *EntryAssign {
	return &EntryAssign{
		message: message{
			mType: MTypeEntryAssign,
		},
		entryName:       entryType.NewString(entryName),
		entry:           entrier,
		entryPersistent: persistent,
		entryID:         id,
		entrySN:         sn,
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (ea *EntryAssign) MarshalMessage(writer io.Writer) error {
	flags := byte(0x00)
	if ea.entryPersistent {
		flags = flags | flagPersistantMask
	}
	_, err := writer.Write([]byte{ea.Type().Byte()})
	if err != nil {
		return err
	}
	err = ea.entryName.MarshalEntry(writer)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{ea.entry.Type()})
	if err != nil {
		return err
	}
	_, err = writer.Write(ea.entryID[:])
	if err != nil {
		return err
	}
	_, err = writer.Write(ea.entrySN[:])
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{flags})
	if err != nil {
		return err
	}
	err = ea.entry.MarshalEntry(writer)
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (ea *EntryAssign) UnmarshalMessage(reader io.Reader) error {
	ea.mType = MTypeEntryAssign
	name := new(entryType.String)
	typeBuf := make([]byte, 1)
	idBuf := make([]byte, 2)
	snBuf := make([]byte, 2)
	flagBuf := make([]byte, 1)

	err := name.UnmarshalEntry(reader)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, typeBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, idBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, snBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, flagBuf)
	if err != nil {
		return err
	}
	ent, err := entryType.Unmarshal(typeBuf[0], reader)
	if err != nil {
		return err
	}

	ea.entryName = name
	ea.entryPersistent = flagBuf[0]&flagPersistantMask == flagPersistantMask
	ea.entry = ent
	copy(ea.entryID[:], idBuf)
	copy(ea.entrySN[:], snBuf)
	return nil
}

func (ea EntryAssign) GetName() string {
	return ea.entryName.String()
}

func (ea EntryAssign) GetEntry() entryType.Entrier {
	return ea.entry
}

//NewEntryID is the ID of an element when a client is creating a new entry.
// (Since go doesn't support constant arrays as of 1.16)
func GetNewEntryID() [2]byte {
	return [2]byte{0xff, 0xff}
}

//NewEntrySN is the sequence number of an element when a client is creating a new entry.
// (Since go doesn't support constant arrays as of 1.16)
func GetNewEntrySN() [2]byte {
	return [2]byte{0x00, 0x00}
}
