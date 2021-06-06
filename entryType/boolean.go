// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"fmt"
	"io"
)

const (
	booleanTrue  byte = 0x01
	booleanFalse byte = 0x00
)

//Boolean is a Network Table Entry that holds the value of type boolean.
type Boolean struct {
	entry
	value bool
}

func (b Boolean) String() string {
	return fmt.Sprintf("%t", b.value)
}

//NewBoolean creates a new instance of a Boolean entry
func NewBoolean(value bool) *Boolean {
	return &Boolean{
		entry: entry{
			eType: ETypeBoolean,
		},
		value: value,
	}
}

//GetValue gets the value of the entry.
func (b *Boolean) GetValue() bool {
	return b.value
}

//SetValue sets the value of the entry.
func (b *Boolean) SetValue(val bool) {
	b.value = val
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (b *Boolean) MarshalEntry(writer io.Writer) error {
	val := booleanFalse
	if b.value {
		val = val | booleanTrue
	}
	_, err := writer.Write([]byte{val})
	return err
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (b *Boolean) UnmarshalEntry(reader io.Reader) error {
	b.eType = ETypeBoolean
	buf := make([]byte, 1)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	b.value = buf[0] == booleanTrue
	return nil
}
