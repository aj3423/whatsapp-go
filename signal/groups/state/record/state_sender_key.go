package record

import (
	"wa/signal/ecc"
	"wa/signal/groups/ratchet"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"

	"google.golang.org/protobuf/proto"
)

const maxMessageKeys = 2000

func NewSenderKeyStateFromBytes(data []byte) (*SenderKeyState, error) {
	p := &pb.SenderKeyStateStructure{}
	e := proto.Unmarshal(data, p)
	if e != nil {
		return nil, e
	}

	return NewSenderKeyStateFromStructure(p)
}

// NewSenderKeyState returns a new SenderKeyState.
func NewSenderKeyState(
	keyID uint32,
	iteration uint32,
	chainKey []byte,
	signatureKey *ecc.ECKeyPair,
) *SenderKeyState {

	return &SenderKeyState{
		keys:           make([]*ratchet.SenderMessageKey, 0, maxMessageKeys/2),
		keyID:          keyID,
		senderChainKey: ratchet.NewSenderChainKey(iteration, chainKey),
		signingKeyPair: signatureKey,
	}
}

// NewSenderKeyStateFromPublicKey returns a new SenderKeyState with the given publicKey.
func NewSenderKeyStateFromPublicKey(
	keyID uint32,
	iteration uint32,
	chainKey []byte,
	signatureKey ecc.ECPublicKeyable,
) *SenderKeyState {

	keyPair := ecc.NewECKeyPair(signatureKey, nil)

	return &SenderKeyState{
		keys:           make([]*ratchet.SenderMessageKey, 0, maxMessageKeys/2),
		keyID:          keyID,
		senderChainKey: ratchet.NewSenderChainKey(iteration, chainKey),
		signingKeyPair: keyPair,
	}
}

// NewSenderKeyStateFromStructure will return a new session state with the
// given state structure. This structure is given back from an
// implementation of the sender key state serializer.
func NewSenderKeyStateFromStructure(structure *pb.SenderKeyStateStructure) (*SenderKeyState, error) {

	// Convert our ecc keys from bytes into object form.
	signingKeyPublic, err := ecc.DecodePoint(structure.GetSenderSigningKey().GetPublic(), 0)
	if err != nil {
		return nil, err
	}
	signingKeyPrivate := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.GetSenderSigningKey().GetPrivate()))

	// Build our sender message keys from structure
	senderMessageKeys := make([]*ratchet.SenderMessageKey, len(structure.GetSenderMessageKeys()))
	for i, k := range structure.GetSenderMessageKeys() {
		m, e := ratchet.NewSenderMessageKey(k.GetIteration(), k.GetSeed())
		if e != nil {
			return nil, e
		}
		senderMessageKeys[i] = m
	}

	// Build our state object.
	state := &SenderKeyState{
		keys:  senderMessageKeys,
		keyID: structure.GetSenderKeyId(),
		senderChainKey: &ratchet.SenderChainKey{
			P: structure.GetSenderChainKey(),
		},
		signingKeyPair: ecc.NewECKeyPair(signingKeyPublic, signingKeyPrivate),
	}

	return state, nil
}

// SenderKeyState is a structure for maintaining a senderkey session state.
type SenderKeyState struct {
	keyID          uint32
	senderChainKey *ratchet.SenderChainKey
	signingKeyPair *ecc.ECKeyPair
	keys           []*ratchet.SenderMessageKey
}

// SigningKey returns the signing key pair of the sender key state.
func (k *SenderKeyState) SigningKey() *ecc.ECKeyPair {
	return k.signingKeyPair
}

// SenderChainKey returns the sender chain key of the state.
func (k *SenderKeyState) SenderChainKey() *ratchet.SenderChainKey {
	return k.senderChainKey
}

// KeyID returns the state's key id.
func (k *SenderKeyState) KeyID() uint32 {
	return k.keyID
}

// HasSenderMessageKey will return true if the state has a key with the
// given iteration.
func (k *SenderKeyState) HasSenderMessageKey(iteration uint32) bool {
	for i := 0; i < len(k.keys); i++ {
		if k.keys[i].Iteration() == iteration {
			return true
		}
	}
	return false
}

// AddSenderMessageKey will add the given sender message key to the state.
func (k *SenderKeyState) AddSenderMessageKey(senderMsgKey *ratchet.SenderMessageKey) {
	k.keys = append(k.keys, senderMsgKey)

	if len(k.keys) > maxMessageKeys {
		k.keys = k.keys[1:]
	}
}

// SetSenderChainKey will set the state's sender chain key with the given key.
func (k *SenderKeyState) SetSenderChainKey(senderChainKey *ratchet.SenderChainKey) {
	k.senderChainKey = senderChainKey
}

// RemoveSenderMessageKey will remove the key in this state with the given iteration number.
func (k *SenderKeyState) RemoveSenderMessageKey(iteration uint32) *ratchet.SenderMessageKey {
	for i := 0; i < len(k.keys); i++ {
		if k.keys[i].Iteration() == iteration {
			removed := k.keys[i]
			k.keys = append(k.keys[0:i], k.keys[i+1:]...)
			return removed
		}
	}

	return nil
}

// Serialize will return the state as bytes using the given serializer.
func (k *SenderKeyState) Serialize() []byte {
	stru := k.structure()
	bs, e := proto.Marshal(stru)
	if e != nil {
		return nil
	}
	return bs
}

// structure will return a serializable structure of the
// the given state so it can be persistently stored.
func (k *SenderKeyState) structure() *pb.SenderKeyStateStructure {
	p := &pb.SenderKeyStateStructure{
		SenderKeyId:    &k.keyID,
		SenderChainKey: k.senderChainKey.P,
	}
	p.SenderSigningKey = &pb.SenderKeyStateStructure_SenderSigningKey{
		Public: k.signingKeyPair.PublicKey().Serialize(),
	}

	if k.signingKeyPair.PrivateKey() != nil {
		p.SenderSigningKey.Private = bytehelper.ArrayToSlice(k.signingKeyPair.PrivateKey().Serialize())
	}
	// Convert our sender message keys into a serializeable structure
	for _, key := range k.keys {
		p.SenderMessageKeys = append(p.SenderMessageKeys, key.P)
	}
	return p
}
