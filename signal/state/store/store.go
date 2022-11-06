package store

import "wa/signal/groups/state/store"

// SignalProtocol store is an interface that implements the
// methods for all stores needed in the Signal Protocol.
type SignalProtocol interface {
	IdentityKey
	PreKey
	Session
	SignedPreKey
	store.SenderKey
}
