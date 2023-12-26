package util

const IndexInvalid uint16 = 0xFFFF

// Dictionary implements a string dictionary that can be persisted to disk.
type Dictionary struct {
	Indexes  map[string]uint16 // Map of indexes for strings
	backRefs map[uint16]string // Back references
	nextIdx  uint16            // Next index to be assigned
}

// NewDictionary returns a new Dictionary ready for use.
func NewDictionary() *Dictionary {
	return &Dictionary{
		Indexes:  map[string]uint16{},
		backRefs: map[uint16]string{},
	}
}

// Put puts a string into the dictionary and returns the index allocated.
func (d *Dictionary) Put(s string) uint16 {
	idx, found := d.Indexes[s]
	if !found {
		d.Indexes[s] = d.nextIdx
		d.backRefs[d.nextIdx] = s
		d.nextIdx++
		return d.nextIdx - 1
	}
	return idx
}

// Get returns the index associated with the string, or IndexInvalid if the
// string is not in the dictionary.
func (d *Dictionary) Get(s string) uint16 {
	v, found := d.Indexes[s]
	if !found {
		return IndexInvalid
	}
	return v
}

// Lookup returns the string associated with the index.
func (d *Dictionary) Lookup(i uint16) string {
	s, found := d.backRefs[i]
	if !found {
		return ""
	}
	return s
}
