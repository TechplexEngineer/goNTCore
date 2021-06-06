package cache

import (
	"fmt"
	"github.com/technomancers/goNTCore/entryType"
	"github.com/technomancers/goNTCore/message"
	"log"
	"sync"
)

// An entry in the NetworkTable
type TableEntry struct {
	entryName       string
	entryID         uint16 //[2]byte
	entrySN         uint16 //[2]byte
	entryPersistent bool
	entry           entryType.Entrier
}

func (t TableEntry) EntryName() string {
	return t.entryName
}

func (t TableEntry) EntryID() uint16 {
	return t.entryID
}

func (t TableEntry) EntrySN() uint16 {
	return t.entrySN
}

func (t TableEntry) EntryPersistent() bool {
	return t.entryPersistent
}

func (t TableEntry) Entry() entryType.Entrier {
	return t.entry
}

type internalData map[string]*TableEntry

// Our local view of the data that has been shared
type NetworkTable struct {
	data internalData
	lock *sync.RWMutex // many readers, one writer
}

// Create a new instance of a network table
func NewNetworkTable() *NetworkTable {
	return &NetworkTable{
		data: make(internalData),
		lock: &sync.RWMutex{},
	}
}

// Create a new entry in the table
func (o *NetworkTable) Assign(ea *message.EntryAssign) {
	o.lock.Lock() // lock for writing
	defer o.lock.Unlock()
	o.data[ea.GetName()] = &TableEntry{
		entryName:       ea.GetName(),
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
				break
			}
			o.data[k].entrySN = entrySN
			o.data[k].entry = entry

			break
		}
	}
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

// Get the number of entriles in the table
func (o *NetworkTable) NumEntries() int {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	return len(o.data)
}

// Get a single entry from the table whose name exactly matches key
func (o *NetworkTable) GetEntry(key string) (TableEntry, error) {
	o.lock.RLock() //lock for reading
	defer o.lock.RUnlock()
	entry, ok := o.data[key]
	if !ok {
		return TableEntry{}, fmt.Errorf("no entry for key: %s", key)
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
