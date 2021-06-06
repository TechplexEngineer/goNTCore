//// Copyright (c) 2017, Technomancers. All rights reserved.
//// Use of this source code is governed by a BSD-style
//// license that can be found in the LICENSE file.
//
package goNTCore

//
//import (
//	"github.com/technomancers/goNTCore/entryType"
//	"github.com/technomancers/goNTCore/util"
//)
//
////NetworkTabler is the interface implemented by types that can read and write Network Table at a high level.
////type NetworkTabler interface {
////	ContainsKey(key string) bool
////	ContainsTable(key string) bool
////	Delete(key string)
////	DeleteAll()
////	IsPersisted(key string) bool
////	GetKeys() []string
////	GetTable(key string) NetworkTabler
////	GetBoolean(key string, def bool) bool
////	PutBoolean(key string, val bool) bool
////	GetNumber(key string, def float64) float64
////	PutNumber(key string, val float64) bool
////	GetString(key string, def string) string
////	PutString(key string, val string) bool
////	GetRaw(key string, def []byte) []byte
////	PutRaw(key string, val []byte) bool
////	GetBooleanArray(key string, def []bool) []bool
////	PutBooleanArray(key string, val []bool) bool
////	GetNumberArray(key string, def []float64) []float64
////	PutNumberArray(key string, val []float64) bool
////	GetStringArray(key string, def []string) []string
////	PutStringArray(key string, val []string) bool
////}
//
////DataTable implements NetworkTabler based on a backend that needs to exist on create.
//type DataTable struct {
//	data *DataTable
//	root string
//}
//
////NewTable creates a new Network Table that is based off of the data and root passed in.
////Pass in an empty string or "/" for root table.
//func NewTable(d *DataTable, root string) *DataTable {
//	if root == "" {
//		root = "/"
//	}
//	root = util.SanatizeKey(root)
//	return &DataTable{
//		data: d,
//		root: root,
//	}
//}
//
////ContainsKey returns true if the key exist in the table.
//func (t *DataTable) ContainsKey(key string) bool {
//	key = t.getKey(key)
//	return t.data.IsKey(key)
//}
//
////ContainsTable return true if the table exist in the table.
//func (t *DataTable) ContainsTable(key string) bool {
//	key = t.getKey(key)
//	return t.data.IsTable(key)
//}
//
////Delete deletes the given key from the table.
//func (t *DataTable) Delete(key string) {
//	key = t.getKey(key)
//	t.data.DeleteEntry(key) // nolint: errcheck
//}
//
////DeleteAll deletes all keys from the table.
//func (t *DataTable) DeleteAll() {
//	t.data.DeleteAll(t.root) // nolint: errcheck
//}
//
////IsPersisted returns true if the key is to be persisted.
////Returns false if the key does not exist.
//func (t *DataTable) IsPersisted(key string) bool {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return false
//	}
//	return entry.Persistent
//}
//
////GetKeys returns all the keys in the table.
//func (t *DataTable) GetKeys() []string {
//	entries, err := t.data.GetEntries(t.root)
//	if err != nil {
//		return []string{}
//	}
//	return entries
//}
//
////GetTable gets a table with the specified key.
////If table does not exist it creates a new one.
//func (t *DataTable) GetTable(key string) *DataTable {
//	key = t.getKey(key)
//	return NewTable(t.data, key)
//}
//
////GetBoolean gets the value of key as a boolean.
////If the value is not of type boolean it returns the default value passed in.
//func (t *DataTable) GetBoolean(key string, def bool) bool {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	b, ok := entry.Value.(bool)
//	if !ok {
//		return def
//	}
//	return b
//}
//
////PutBoolean puts the boolean value in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutBoolean(key string, val bool) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeBoolean,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetNumber gets the value of key as a float64.
////If the value is not of type float64 it returns the default value passed in.
//func (t *DataTable) GetNumber(key string, def float64) float64 {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	d, ok := entry.Value.(float64)
//	if !ok {
//		return def
//	}
//	return d
//}
//
////PutNumber puts the float64 value in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutNumber(key string, val float64) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeDouble,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetString gets the value of key as a string.
////If the value is not of type string it returns the default value passed in.
//func (t *DataTable) GetString(key string, def string) string {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	s, ok := entry.Value.(string)
//	if !ok {
//		return def
//	}
//	return s
//}
//
////PutString puts the string value in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutString(key string, val string) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeString,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetRaw gets the value of key as a slice of bytes.
////If the value is not of type byte slice it returns the default value passed in.
//func (t *DataTable) GetRaw(key string, def []byte) []byte {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	r, ok := entry.Value.([]byte)
//	if !ok {
//		return def
//	}
//	return r
//}
//
////PutRaw puts the taw value in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutRaw(key string, val []byte) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeRawData,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetBooleanArray gets the value of key as a slice of booleans.
////If the value is not of type boolean slice it returns the default value passed in.
//func (t *DataTable) GetBooleanArray(key string, def []bool) []bool {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	ba, ok := entry.Value.([]bool)
//	if !ok {
//		return def
//	}
//	return ba
//}
//
////PutBooleanArray puts the slice of boolean in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutBooleanArray(key string, val []bool) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeBooleanArray,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetNumberArray gets the value of key as a slice of float64s.
////If the value is not of type float64 slice it returns the default value passed in.
//func (t *DataTable) GetNumberArray(key string, def []float64) []float64 {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	da, ok := entry.Value.([]float64)
//	if !ok {
//		return def
//	}
//	return da
//}
//
////PutNumberArray puts the slice of float64 in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutNumberArray(key string, val []float64) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeDoubleArray,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
////GetStringArray gets the value of key as a slice of strings.
////If the value is not of type string slice it returns the default value passed in.
//func (t *DataTable) GetStringArray(key string, def []string) []string {
//	key = t.getKey(key)
//	entry, err := t.data.Entry(key)
//	if err != nil {
//		return def
//	}
//	sa, ok := entry.Value.([]string)
//	if !ok {
//		return def
//	}
//	return sa
//}
//
////PutStringArray puts the slice of string in the table.
////If value exist it updates it.
////If value doesn't exist it will add it.
////Returns false if key exist of a different type.
//func (t *DataTable) PutStringArray(key string, val []string) bool {
//	entry := &Entry{
//		Key:   key,
//		Type:  entryType.ETypeStringArray,
//		Value: val,
//	}
//	if err := t.data.PutEntry(entry); err != nil {
//		return false
//	}
//	return true
//}
//
//func (t *DataTable) getKey(key string) string {
//	return util.KeyJoin(t.root, key)
//}
//
//func (t *DataTable) Entry(key string) (Entry, error) {
//
//}
//
//func (t *DataTable) GetEntries(root string) ([]string, error) {
//
//}
//
//func (t *DataTable) PutEntry(entry *Entry) error {
//
//}
//
//func (t *DataTable) IsKey(key string) bool {
//
//}
//
//func (t *DataTable) IsTable(key string) bool {
//
//}
//
//func (t *DataTable) DeleteEntry(key string) {
//
//}
