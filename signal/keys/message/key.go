// Package message provides a structure for message keys, which are symmetric
// keys used for the encryption/decryption of Signal messages.
package message

import (
	"wa/signal/pb"

	"google.golang.org/protobuf/proto"
)

// DerivedSecretsSize is the size of the derived secrets for message keys.
const DerivedSecretsSize = 80

// CipherKeyLength is the length of the actual cipher key used for messages.
const CipherKeyLength = 32

// MacKeyLength is the length of the message authentication code in bytes.
const MacKeyLength = 32

// IVLength is the length of the initialization vector in bytes.
const IVLength = 16

// KdfSalt is used as the Salt for message keys to derive secrets using a Key Derivation Function
const KdfSalt string = "WhisperMessageKeys"

// NewKeys returns a new message keys structure with the given cipherKey, mac, iv, and index.
func NewKeys(cipherKey, macKey, iv []byte, index uint32) *Keys {
	messageKeys := Keys{
		cipherKey: cipherKey,
		macKey:    macKey,
		iv:        iv,
		index:     index,
	}

	return &messageKeys
}

// NewKeysFromStruct will return a new message keys object from the
// given serializeable structure.
func NewKeysFromStruct(structure *pb.SessionStructure_Chain_MessageKey) *Keys {
	return NewKeys(
		structure.CipherKey,
		structure.MacKey,
		structure.Iv,
		structure.GetIndex(),
	)
}

// NewStructFromKeys returns a serializeable structure of message keys.
func NewStructFromKeys(keys *Keys) *pb.SessionStructure_Chain_MessageKey {
	return &pb.SessionStructure_Chain_MessageKey{
		CipherKey: keys.cipherKey,
		MacKey:    keys.macKey,
		Iv:        keys.iv,
		Index:     proto.Uint32(keys.index),
	}
}

// Keys is a structure to hold all the keys for a single MessageKey, including the
// cipherKey, mac, iv, and index of the chain key. MessageKeys are used to
// encrypt individual messages.
type Keys struct {
	cipherKey []byte
	macKey    []byte
	iv        []byte
	index     uint32
}

// CipherKey is the key used to produce ciphertext.
func (k *Keys) CipherKey() []byte {
	return k.cipherKey
}

// MacKey returns the message's message authentication code.
func (k *Keys) MacKey() []byte {
	return k.macKey
}

// Iv returns the message keys' initialization vector. The IV is a fixed-size input
// to a cryptographic primitive.
func (k *Keys) Iv() []byte {
	return k.iv
}

// Index returns the number of times the chain key has been put through a key derivation
// function to generate this message key.
func (k *Keys) Index() uint32 {
	return k.index
}
