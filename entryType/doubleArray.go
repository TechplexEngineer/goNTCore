// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"encoding/json"
	"fmt"
	"io"
)

//DoubleArray is a Network Table Entry that holds the value of type Array of Doubles.
type DoubleArray struct {
	entry
	value []*Double
}

func (da DoubleArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(da.value)
}

func (da DoubleArray) String() string {
	return fmt.Sprintf("%v", da.value)
}

//NewDoubleArray creates an instance of DoubleArray.
func NewDoubleArray(value []float64) *DoubleArray {
	da := &DoubleArray{
		entry: entry{
			eType: ETypeDoubleArray,
		},
	}
	for _, d := range value {
		double := NewDouble(d)
		da.value = append(da.value, double)
	}
	return da
}

//GetValue gets the value of the string.
func (da *DoubleArray) GetValue() []float64 {
	var out []float64
	for _, d := range da.value {
		out = append(out, d.GetValue())
	}
	return out
}

//SetValue sets the value of the entry.
func (da *DoubleArray) SetValue(val []float64) {
	// clear out the previous values
	da.value = make([]*Double, 1)

	for _, v := range val {
		da.value = append(da.value, NewDouble(v))
	}
}

//MarshalEntry implements Marshaler for Network Table Entry.
func (da *DoubleArray) MarshalEntry(writer io.Writer) error {
	lenArray := byte(len(da.value))
	_, err := writer.Write([]byte{lenArray})
	if err != nil {
		return err
	}
	for _, d := range da.value {
		err = d.MarshalEntry(writer)
		if err != nil {
			return err
		}
	}
	return nil
}

//UnmarshalEntry implements Unmarshaler for Network Table Entry.
func (da *DoubleArray) UnmarshalEntry(reader io.Reader) error {
	da.eType = ETypeDoubleArray
	lenBuf := make([]byte, 1)
	_, err := io.ReadFull(reader, lenBuf)
	if err != nil {
		return err
	}
	numEle := int(lenBuf[0])
	for i := 0; i < numEle; i++ {
		double := new(Double)
		err = double.UnmarshalEntry(reader)
		if err != nil {
			return err
		}
		da.value = append(da.value, double)
	}
	return nil
}
