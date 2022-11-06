package ratchet

import (
	"wa/signal/kdf"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"
)

// KdfInfo is optional bytes to include in deriving secrets with KDF.
const KdfInfo string = "WhisperGroup"

// NewSenderMessageKey will return a new sender message key using the given
// iteration and seed.
func NewSenderMessageKey(iteration uint32, seed []byte) (*SenderMessageKey, error) {
	derivative, err := kdf.DeriveSecrets(seed, nil, []byte(KdfInfo), 48)
	if err != nil {
		return nil, err
	}

	// Split our derived secrets into 2 parts
	parts := bytehelper.Split(derivative, 16, 32)

	// Build the message key.
	senderKeyMessage := &SenderMessageKey{
		P: &pb.SenderKeyStateStructure_SenderMessageKey{
			Iteration: &iteration,
			Seed:      seed,
		},
		iv:        parts[0],
		cipherKey: parts[1],
	}

	return senderKeyMessage, nil
}

// SenderMessageKey is a structure for sender message keys used in group messaging.
type SenderMessageKey struct {
	P *pb.SenderKeyStateStructure_SenderMessageKey

	iv        []byte
	cipherKey []byte
}

// Iteration will return the sender message key's iteration.
func (k *SenderMessageKey) Iteration() uint32 {
	return k.P.GetIteration()
}

// Iv will return the sender message key's initialization vector.
func (k *SenderMessageKey) Iv() []byte {
	return k.iv
}

// CipherKey will return the key in bytes.
func (k *SenderMessageKey) CipherKey() []byte {
	return k.cipherKey
}

// Seed will return the sender message key's seed.
func (k *SenderMessageKey) Seed() []byte {
	return k.P.GetSeed()
}
