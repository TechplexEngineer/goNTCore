// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"fmt"
	"io"

	"github.com/techplexengineer/gontcore/util"
)

//RawData is any type that is not covered in the protocole.
//It is recommended to add a byte or two to the beginning to describe what type it is.
type RawData struct {
	entry
	data []byte
}

func (rd RawData) MarshalJSON() ([]byte, error) {
	return rd.data, nil
}

func (rd RawData) String() string {
	return fmt.Sprintf("%v", rd.data)
}

//NewRawData creates and instance of RawData.
func NewRawData(data []byte) *RawData {
	return &RawData{
		entry: entry{
			eType: ETypeRawData,
		},
		data: data,
	}
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (rd *RawData) MarshalEntry(writer io.Writer) error {
	err := util.EncodeULeb128(uint32(len(rd.data)), writer)
	if err != nil {
		return err
	}
	_, err = writer.Write(rd.data)
	return err
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (rd *RawData) UnmarshalEntry(reader io.Reader) error {
	rd.eType = ETypeRawData
	dataLength, err := util.DecodeULeb128(reader)
	if err != nil {
		return err
	}
	buf := make([]byte, dataLength)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return err
	}
	rd.data = buf
	return nil
}

//GetValue gets the value of the entry.
func (rd *RawData) GetValue() []byte {
	return rd.data
}

//SetValue sets the value of the entry.
func (b *RawData) SetValue(val []byte) {
	b.data = val
}
