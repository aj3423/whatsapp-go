package ecc

// Package curve25519sign implements a signature scheme based on Curve25519 keys.
// See https://moderncrypto.org/mail-archive/curves/2014/000205.html for details.

import (
	"crypto/ed25519"

	"wa/signal/ecc/edwards25519"
)

/*

This is due to different ed25519 private key formats. An ed25519 key starts out as a 32 byte seed. This seed is hashed
with SHA512 to produce 64 bytes (a couple of bits are flipped too). The first 32 bytes of these are used to generate
the public key (which is also 32 bytes), and the last 32 bytes are used in the generation of the signature.

The Golang private key format is the 32 byte seed concatenated with the 32 byte public key.
The private keys we are using are the 64 byte result of the hash
(or possibly just 64 random bytes that are used the same way as the hash result).

Since it’s not possible to reverse the hash, you can’t convert the 64 byte keys to a format that the Golang API
will accept.

You can produce a version of the Golang lib based on the existing package.

The following code depends on the internal package golang.org/x/crypto/ed25519/internal/edwards25519,
so if you want to use it you will need to copy that package out so that it is available to you code.
It’s also very “rough and ready”, I’ve basically just copied the chunks of code needed
from the existing code to get this to work.

*/

// sign signs the message with privateKey and returns a signature as a byte slice.
func sign(privateKey *[32]byte, message []byte, random [64]byte) *[64]byte {
	panic(`replaced by xed25519.Sign`)
	return nil
}

// verify checks whether the message has a valid signature.
func verify(publicKey [32]byte, message []byte, signature *[64]byte) bool {

	publicKey[31] &= 0x7F

	/* Convert the Curve25519 public key into an Ed25519 public key.  In
	particular, convert Curve25519's "montgomery" x-coordinate into an
	Ed25519 "edwards" y-coordinate:

	ed_y = (mont_x - 1) / (mont_x + 1)

	NOTE: mont_x=-1 is converted to ed_y=0 since fe_invert is mod-exp

	Then move the sign bit into the pubkey from the signature.
	*/

	var edY, one, montX, montXMinusOne, montXPlusOne edwards25519.FieldElement
	edwards25519.FeFromBytes(&montX, &publicKey)
	edwards25519.FeOne(&one)
	edwards25519.FeSub(&montXMinusOne, &montX, &one)
	edwards25519.FeAdd(&montXPlusOne, &montX, &one)
	edwards25519.FeInvert(&montXPlusOne, &montXPlusOne)
	edwards25519.FeMul(&edY, &montXMinusOne, &montXPlusOne)

	var A_ed [32]byte
	edwards25519.FeToBytes(&A_ed, &edY)

	A_ed[31] |= signature[63] & 0x80
	signature[63] &= 0x7F

	sig := *signature

	return ed25519.Verify(A_ed[:], message, sig[:])
}

// sign32 signs the message with privateKey and returns a signature as a byte slice.
func sign32(privateKey [32]byte, message []byte, random [64]byte) (signature [64]byte) {
	pk := [64]byte{}
	copy(pk[:], privateKey[:32])
	sig := ed25519.Sign(pk[:], message)
	copy(signature[:], sig[:])

	return signature
}

// verify32 checks whether the message has a valid signature.
func verify32(publicKey [32]byte, message []byte, signature [64]byte) bool {
	return ed25519.Verify(publicKey[:], message, signature[:])
}
