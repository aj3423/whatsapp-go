package ecc

import "wa/signal/util/bytehelper"

// NewDjbECPublicKey creates a new Curve25519 public key with the given bytes.
func NewDjbECPublicKey(publicKey []byte) *DjbECPublicKey {
	len_ := len(publicKey)

	var pub32 []byte

	if len_ == 32 {
		pub32 = publicKey
	} else if len_ == 33 && publicKey[0] == DjbType {
		pub32 = publicKey[1:]
	} else {
		return nil
	}
	key := DjbECPublicKey{
		publicKey: bytehelper.SliceToArray(pub32),
	}
	return &key
}

// DjbECPublicKey implements the ECPublicKey interface and uses Curve25519.
type DjbECPublicKey struct {
	publicKey [32]byte
}

// PublicKey returns the EC public key as a byte array.
func (d *DjbECPublicKey) PublicKey() [32]byte {
	return d.publicKey
}

// Serialize returns the public key prepended by the DjbType value.
func (d *DjbECPublicKey) Serialize() []byte {
	return append([]byte{DjbType}, d.publicKey[:]...)
}

// Type returns the DjbType value.
func (d *DjbECPublicKey) Type() int {
	return DjbType
}
