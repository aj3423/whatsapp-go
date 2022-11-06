package ecc

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/curve25519"

	"ahex"
	"algo/xed25519"
	"wa/signal/util/bytehelper"
)

// DjbType is the Diffie-Hellman curve type (curve25519) created by D. J. Bernstein.
const DjbType = 0x05

// DecodePoint will take the given bytes and offset and return an ECPublicKeyable object.
// This is used to check the byte at the given offset in the byte array for a special
// "type" byte that will determine the key type. Currently only DJB EC keys are supported.
func DecodePoint(bytes []byte, offset int) (ECPublicKeyable, error) {
	ret := NewDjbECPublicKey(bytes[offset:])
	if ret == nil {
		return nil, errors.New("Fail Decode Point: " + ahex.Enc(bytes))
	}
	return ret, nil
}

func CreateKeyPair(privateKey []byte) *ECKeyPair {
	var private, public [32]byte
	copy(private[:], privateKey)

	private[0] &= 248
	private[31] &= 127
	private[31] |= 64

	curve25519.ScalarBaseMult(&public, &private)

	// Put data into our keypair struct
	djbECPub := NewDjbECPublicKey(bytehelper.ArrayToSlice(public))
	djbECPriv := NewDjbECPrivateKey(private)
	keypair := NewECKeyPair(djbECPub, djbECPriv)

	return keypair
}

// GenerateKeyPair returns an EC Key Pair.
func GenerateKeyPair() (*ECKeyPair, error) {
	// Get cryptographically secure random numbers.
	random := rand.Reader

	// Create a byte array for our public and private keys.
	var private, public [32]byte

	// Generate some random data
	_, err := io.ReadFull(random, private[:])
	if err != nil {
		return nil, err
	}

	// Documented at: http://cr.yp.to/ecdh.html
	private[0] &= 248
	private[31] &= 127
	private[31] |= 64

	curve25519.ScalarBaseMult(&public, &private)

	// Put data into our keypair struct
	djbECPub := NewDjbECPublicKey(bytehelper.ArrayToSlice(public))
	djbECPriv := NewDjbECPrivateKey(private)
	keypair := NewECKeyPair(djbECPub, djbECPriv)

	return keypair, nil
}

func VerifySignature(signingKey ECPublicKeyable, message []byte, signature [64]byte) bool {
	publicKey := signingKey.PublicKey()
	valid := verify(publicKey, message, &signature)
	return valid
}

//var rands = []string{}

//func SetRands(r []string) {
//rands = r
//}

func CalculateSignature(signingKey ECPrivateKeyable, message []byte) ([64]byte, error) {
	privateKey := signingKey.Serialize()

	// for test
	//var random []byte
	//if len(rands) > 0 {
	//random = ahex.Dec(rands[0])
	//rands = rands[1:]
	//} else {
	//random = arand.Bytes(64)
	//}
	//sig, e := xed25519.DoSign(privateKey[:], message, random)

	sig, e := xed25519.Sign(privateKey[:], message)
	if e != nil {
		return [64]byte{}, e
	}
	return bytehelper.SliceToArray64(sig), nil
}
