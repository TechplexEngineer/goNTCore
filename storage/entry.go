package storage

import "github.com/technomancers/goNTCore/entryType"

// An entry in the NetworkTable
// implements the ITableEntry interface
type StorageEntry struct {
	entryName       string
	entryID         uint16 //[2]byte
	entrySN         uint16 //[2]byte
	entryPersistent bool
	entry           entryType.Entrier
}

func (t StorageEntry) GetName() string {
	return t.entryName
}

func (t StorageEntry) GetID() uint16 {
	return t.entryID
}

func (t StorageEntry) GetSN() uint16 {
	return t.entrySN
}

func (t StorageEntry) IsPersistent() bool {
	return t.entryPersistent
}

func (t StorageEntry) GetEntry() entryType.Entrier {
	return t.entry
}
