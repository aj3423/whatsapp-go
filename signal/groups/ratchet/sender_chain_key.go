package ratchet

import (
	"crypto/hmac"
	"crypto/sha256"

	"wa/signal/pb"
)

var messageKeySeed = []byte{0x01}
var chainKeySeed = []byte{0x02}

// NewSenderChainKey will return a new SenderChainKey.
func NewSenderChainKey(iteration uint32, chainKey []byte) *SenderChainKey {
	return &SenderChainKey{
		P: &pb.SenderKeyStateStructure_SenderChainKey{
			Iteration: &iteration,
			Seed:      chainKey,
		},
	}
}

type SenderChainKey struct {
	P *pb.SenderKeyStateStructure_SenderChainKey
}

func (k *SenderChainKey) Iteration() uint32 {
	return k.P.GetIteration()
}

func (k *SenderChainKey) SenderMessageKey() (*SenderMessageKey, error) {
	return NewSenderMessageKey(k.P.GetIteration(), k.getDerivative(messageKeySeed, k.P.GetSeed()))
}

func (k *SenderChainKey) Next() *SenderChainKey {
	return NewSenderChainKey(k.P.GetIteration()+1, k.getDerivative(chainKeySeed, k.P.GetSeed()))
}

func (k *SenderChainKey) Seed() []byte {
	return k.P.GetSeed()
}

func (k *SenderChainKey) getDerivative(seed []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key[:])
	mac.Write(seed)

	return mac.Sum(nil)
}
