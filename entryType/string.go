// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"fmt"
	"github.com/techplexengineer/gontcore/util"
)
import "io"

//String is a Network Table Entry that holds the value of type string.
type String struct {
	entry
	value string
}

func (o String) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, o.value)), nil
}

func (o String) String() string {
	return o.value
}

//NewString creates a new Network Table Entry of type string.
func NewString(value string) *String {
	return &String{
		entry: entry{
			eType: ETypeString,
		},
		value: value,
	}
}

//GetValue gets the value of the string.
func (s *String) GetValue() string {
	return s.value
}

//SetValue sets the value of the entry.
func (d *String) SetValue(val string) {
	d.value = val
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (s *String) MarshalEntry(writer io.Writer) error {
	valueBytes := []byte(s.value)
	err := util.EncodeULeb128(uint32(len(valueBytes)), writer)
	if err != nil {
		return err
	}
	_, err = writer.Write(valueBytes)
	return err
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (s *String) UnmarshalEntry(reader io.Reader) error {
	s.eType = ETypeString
	lengthString, err := util.DecodeULeb128(reader)
	if err != nil {
		return err
	}
	buf := make([]byte, lengthString)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	s.value = string(buf)
	return nil
}
