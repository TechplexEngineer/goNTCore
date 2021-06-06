// Copyright (c) 2017, Technomancers. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package entryType

import (
	"errors"
	"io"
)

type EntryType byte

func (o EntryType) String() string {

	switch o {
	case ETypeBoolean:
		return "Boolean"
	case ETypeDouble:
		return "Double"
	case ETypeString:
		return "String"
	case ETypeRawData:
		return "RawData"
	case ETypeBooleanArray:
		return "BooleanArray"
	case ETypeDoubleArray:
		return "DoubleArray"
	case ETypeStringArray:
		return "StringArray"
	case ETypeRPCDef:
		return "RPCDef"
	default:
		return "Unknown Entry Type"
	}
}

func (o EntryType) Byte() byte {
	return byte(o)
}

const (
	ETypeBoolean      EntryType = 0x00
	ETypeDouble       EntryType = 0x01
	ETypeString       EntryType = 0x02
	ETypeRawData      EntryType = 0x03
	ETypeBooleanArray EntryType = 0x10
	ETypeDoubleArray  EntryType = 0x11
	ETypeStringArray  EntryType = 0x12
	ETypeRPCDef       EntryType = 0x20
)

type entry struct {
	eType EntryType
}

func (e entry) Type() EntryType {
	return e.eType
}

//Marshaler is the interface implemented by types that can marshal themselves into valid Network Table Entry Value.
type Marshaler interface {
	MarshalEntry(io.Writer) error
}

//Unmarshaler is the interface implemented by types that can unmarshal a Network Table Entry Value of themselves.
type Unmarshaler interface {
	UnmarshalEntry(io.Reader) error
}

//Entrier is the interface implemented by types that can be an Entry in the Network Tables.
type Entrier interface {
	Type() EntryType
	Marshaler
	Unmarshaler
	String() string
}

//Unmarshal takes the type passed in and tries to unmarshal the next bytes from reader based on the type.
//It returns an instance entry.
func Unmarshal(t byte, reader io.Reader) (Entrier, error) {
	var ent Entrier
	switch EntryType(t) {
	case ETypeBoolean:
		ent = new(Boolean)
	case ETypeDouble:
		ent = new(Double)
	case ETypeString:
		ent = new(String)
	case ETypeRawData:
		ent = new(RawData)
	case ETypeBooleanArray:
		ent = new(BooleanArray)
	case ETypeDoubleArray:
		ent = new(DoubleArray)
	case ETypeStringArray:
		ent = new(StringArray)
	default:
		return nil, errors.New("Unmarshal Entry: Could not find appropropriate type")
	}
	err := ent.UnmarshalEntry(reader)
	return ent, err
}
