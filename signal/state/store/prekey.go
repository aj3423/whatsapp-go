package store

import "wa/signal/state/record"

// PreKey store is an interface describing the local storage
// of PreKeyRecords
type PreKey interface {
	// Load a local PreKeyRecord
	LoadPreKey(preKeyID uint32) (*record.PreKey, error)

	// Store a local PreKeyRecord
	StorePreKey(preKeyID uint32, preKeyRecord *record.PreKey) error

	// Check to see if the store contains a PreKeyRecord
	ContainsPreKey(preKeyID uint32) bool

	// Delete a PreKeyRecord from local storage.
	RemovePreKey(preKeyID uint32)
}
