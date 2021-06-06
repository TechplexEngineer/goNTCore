// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"fmt"
	"io"
)

//StringArray is a Network Table Entry that holds the value of type Array of Strings.
type StringArray struct {
	entry
	value []*String
}

func (sa StringArray) String() string {
	return fmt.Sprintf("%v", sa.value)
}

//NewStringArray creates an instance of StringArray.
func NewStringArray(value []string) *StringArray {
	sa := &StringArray{
		entry: entry{
			eType: ETypeStringArray,
		},
	}
	for _, s := range value {
		st := NewString(s)
		sa.value = append(sa.value, st)
	}
	return sa
}

//GetValue gets the value of the entry.
func (sa *StringArray) GetValue() []string {
	var out []string
	for _, s := range sa.value {
		out = append(out, s.GetValue())
	}
	return out
}

//SetValue sets the value of the entry.
func (da *StringArray) SetValue(val []string) {
	// clear out the previous values
	da.value = make([]*String, 1)

	for _, v := range val {
		da.value = append(da.value, NewString(v))
	}
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (sa *StringArray) MarshalEntry(writer io.Writer) error {
	lenArray := byte(len(sa.value))
	_, err := writer.Write([]byte{lenArray})
	if err != nil {
		return err
	}
	for _, s := range sa.value {
		err = s.MarshalEntry(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (sa *StringArray) UnmarshalEntry(reader io.Reader) error {
	sa.eType = ETypeStringArray
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(reader, lenBuf)
	if err != nil {
		return err
	}
	numEle := int(lenBuf[0])
	for i := 0; i < numEle; i++ {
		st := new(String)
		err = st.UnmarshalEntry(reader)
		if err != nil {
			return err
		}
		sa.value = append(sa.value, st)
	}
	return nil
}
