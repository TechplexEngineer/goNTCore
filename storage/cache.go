package storage

import (
	"encoding/json"
	"fmt"
	"github.com/techplexengineer/gontcore/entryType"
	"github.com/techplexengineer/gontcore/message"
	"github.com/techplexengineer/gontcore/util"
	"log"
	"sync"
)

type internalData map[string]*StorageEntry

// Our local view of the data that has been shared
type NetworkTable struct {
	data internalData
	lock *sync.RWMutex // many readers, one writer
}

// Create a new instance of a network table
func NewNetworkTable() *NetworkTable {
	return &NetworkTable{
		data: make(internalData),
		lock: new(sync.RWMutex),
	}
}

// Create a new entry in the table
func (o *NetworkTable) Assign(ea *message.EntryAssign) {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()
	o.data[ea.GetName()] = &StorageEntry{
		entryName:       util.SanitizeKey(ea.GetName()),
		entryID:         ea.EntryID(),
		entrySN:         ea.EntrySN(),
		entryPersistent: ea.EntryPersistent(),
		entry:           ea.Entry(),
	}
}

// Update an entry in the table
func (o *NetworkTable) Update(entryId uint16, entrySN uint16, entry entryType.Entrier) {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()

	// @todo Should we optimize for writes by ID or reads by Key... currently optimized for reads by key.
	for k, v := range o.data {
		if v.entryID == entryId {
			if v.entry.Type() != entry.Type() {
				log.Printf("Types differ. Ignoring update")
				return
			}
			o.data[k].entrySN = entrySN
			o.data[k].entry = entry

			return
		}
	}
	log.Printf("entryId %d not found, unable to update", entryId)
}

// Delete an entry from the table whose ID exactly matches entryId
func (o *NetworkTable) DeleteID(entryId uint16) {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()
	for k, v := range o.data {
		if v.entryID == entryId {
			delete(o.data, k)
		}
	}
}

// Remove entry that exactly matches key
func (o *NetworkTable) DeleteKey(key string) {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()
	delete(o.data, key)
}

// remove all entries in the map
func (o *NetworkTable) DeleteAll() {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()
	o.data = make(internalData)
}

// Get the number of entries in the table
func (o *NetworkTable) NumEntries() int {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	return len(o.data)
}

// Get a single entry from the table whose name exactly matches key
func (o *NetworkTable) GetEntry(key string) (StorageEntry, error) {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	entry, ok := o.data[key]
	if !ok {
		return StorageEntry{}, fmt.Errorf("no entry for key: %s", key)
	}
	return *entry, nil
}

// Get a lost of the keys in the table
func (o *NetworkTable) GetKeys() []string {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	res := []string{}
	for k, _ := range o.data {
		res = append(res, k)
	}
	return res
}

// Find the Name (key) for the given entry id
func (o *NetworkTable) IdToName(entryId uint16) (string, error) {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	for k, v := range o.data {
		if v.entryID == entryId {
			return k, nil
		}
	}
	return "", fmt.Errorf("inable to find id %d", entryId)
}

type SnapShotEntry struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	Datatype     string `json:"type"`
	ID           uint16 `json:"id"`
	SN           uint16 `json:"sn"`
	IsPersistent bool   `json:"is_persistent"`
}

func (o *NetworkTable) GetSnapshot() []SnapShotEntry {
	keys := []SnapShotEntry{}
	for k, v := range o.data {

		valueStr := fmt.Sprintf("%#v", v.GetEntry())
		valueByt, err := json.Marshal(v.GetEntry())
		if err == nil {
			valueStr = string(valueByt)
		}

		keys = append(keys, SnapShotEntry{
			Key:          k,
			Value:        valueStr,
			Datatype:     v.GetEntry().Type().String(),
			ID:           v.GetID(),
			SN:           v.GetSN(),
			IsPersistent: v.IsPersistent(),
		})

	}
	return keys

}
