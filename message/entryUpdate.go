// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/techplexengineer/gontcore/entryType"
)

//EntryUpdate is used to tell the network that an entry has been updated.
type EntryUpdate struct {
	message
	entryID [2]byte
	entrySN [2]byte
	entrier entryType.Entrier
}

//implements the stringer interface
func (m EntryUpdate) String() string {
	return fmt.Sprintf("EntryUpdate - ID:%#x SN:%#x value:%s", m.entryID, m.entrySN, m.entrier)
}

func (m EntryUpdate) Entry() entryType.Entrier {
	return m.entrier
}

func (m EntryUpdate) EntrySN() uint16 {
	return binary.BigEndian.Uint16(m.entrySN[:])
}

func (m EntryUpdate) EntryID() uint16 {
	return binary.BigEndian.Uint16(m.entryID[:])
}

//NewEntryUpdate creates a new instance on EntryUpdate.
func NewEntryUpdate(id, sn uint16, entrier entryType.Entrier) *EntryUpdate {
	idBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(idBuf, id)

	snBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(snBuf, sn)

	eu := &EntryUpdate{
		message: message{
			mType: MTypeEntryUpdate,
		},
		//entryID: idBuf,
		//entrySN: snBuf,
		entrier: entrier,
	}

	copy(eu.entryID[:], idBuf)
	copy(eu.entrySN[:], snBuf)

	return eu
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (eu *EntryUpdate) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{eu.Type().Byte()})
	if err != nil {
		return err
	}
	_, err = writer.Write(eu.entryID[:])
	if err != nil {
		return err
	}
	_, err = writer.Write(eu.entrySN[:])
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte{eu.entrier.Type().Byte()})
	if err != nil {
		return err
	}
	err = eu.entrier.MarshalEntry(writer)
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type byte has already been read.
func (eu *EntryUpdate) UnmarshalMessage(reader io.Reader) error {
	eu.mType = MTypeEntryUpdate
	idBuf := make([]byte, 2)
	snBuf := make([]byte, 2)
	typeBuf := make([]byte, 1)

	_, err := io.ReadFull(reader, idBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, snBuf)
	if err != nil {
		return err
	}
	_, err = io.ReadFull(reader, typeBuf)
	if err != nil {
		return err
	}
	ent, err := entryType.Unmarshal(typeBuf[0], reader)
	if err != nil {
		return err
	}

	eu.entrier = ent
	copy(eu.entryID[:], idBuf)
	copy(eu.entrySN[:], snBuf)
	return nil
}
