package identity

import "wa/signal/ecc"

func NewKeyPair(publicKey *Key, privateKey ecc.ECPrivateKeyable) *KeyPair {
	keyPair := KeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}

	return &keyPair
}

//func NewKeyPairFromBytes(priv, pub []byte) *KeyPair {
//}

type KeyPair struct {
	publicKey  *Key
	privateKey ecc.ECPrivateKeyable
}

func (k *KeyPair) PublicKey() *Key {
	return k.publicKey
}

func (k *KeyPair) PrivateKey() ecc.ECPrivateKeyable {
	return k.privateKey
}

//func (k *KeyPair) Serialize() []byte {
//}
