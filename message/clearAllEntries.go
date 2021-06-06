// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message

import (
	"fmt"
	"io"

	"github.com/technomancers/goNTCore/util"
)

var (
	clearAllMagic = [4]byte{0xd0, 0x6c, 0xb2, 0x7a}
)

//ClearAllEntries clears all entries from the network.
type ClearAllEntries struct {
	message
	magic [4]byte
	valid bool
}

//implements the stringer interface
func (m ClearAllEntries) String() string {
	return fmt.Sprintf("ClearAll - magic:%#x valid:%t", m.magic, m.valid)
}

func (m ClearAllEntries) Valid() bool {
	return m.valid
}

func (m ClearAllEntries) Magic() [4]byte {
	return m.magic
}

//NewClearAllEntries creates an instance of clear all entries.
func NewClearAllEntries() *ClearAllEntries {
	return &ClearAllEntries{
		message: message{
			mType: MTypeClearAllEntries,
		},
		magic: clearAllMagic,
		valid: true,
	}
}

//MarshalMessage implements Marshaler for Network Table Messages.
func (cae *ClearAllEntries) MarshalMessage(writer io.Writer) error {
	_, err := writer.Write([]byte{cae.Type().Byte()})
	if err != nil {
		return err
	}
	_, err = writer.Write(cae.magic[:])
	return err
}

//UnmarshalMessage implements Unmarshaler for Network Table Messages and assumes the message type bit has already been read.
func (cae *ClearAllEntries) UnmarshalMessage(reader io.Reader) error {
	cae.mType = MTypeClearAllEntries
	buf := make([]byte, 4)

	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return err
	}

	cae.valid = util.Match(cae.magic[:], clearAllMagic[:])
	return nil
}
