package fingerprint

import "wa/signal/keys/identity"

// FingerprintGenerator is an interface for fingerprint generators.
type FingerprintGenerator interface {
	CreateFor(localStableIdentifier, remoteStableIdentifier string, localIdentityKey, remoteIdentityKey *identity.Key) *Fingerprint
	CreateForMultiple(localStableIdentifier, remoteStableIdentifier string, localIdentityKey, remoteIdentityKey []*identity.Key) *Fingerprint
}
