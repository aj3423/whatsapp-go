package record

import (
	"wa/signal/ecc"
	"wa/signal/kdf"
	"wa/signal/keys/chain"
	"wa/signal/keys/message"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"
)

// NewReceiverChainPair will return a new ReceiverChainPair object.
func NewReceiverChainPair(receiverChain *Chain, index int) *ReceiverChainPair {
	return &ReceiverChainPair{
		ReceiverChain: receiverChain,
		Index:         index,
	}
}

// ReceiverChainPair is a structure for a receiver chain key and index number.
type ReceiverChainPair struct {
	ReceiverChain *Chain
	Index         int
}

// NewChain returns a new Chain structure for SessionState.
func NewChain(senderRatchetKeyPair *ecc.ECKeyPair, chainKey *chain.Key,
	messageKeys []*message.Keys) *Chain {

	return &Chain{
		senderRatchetKeyPair: senderRatchetKeyPair,
		chainKey:             chainKey,
		messageKeys:          messageKeys,
	}
}

// NewChainFromStructure will return a new Chain with the given
// chain structure.
func NewChainFromStructure(structure *pb.SessionStructure_Chain) (*Chain, error) {
	// Alias to SliceToArray
	getArray := bytehelper.SliceToArray

	// Build the sender ratchet key from bytes.
	senderRatchetKeyPublic, err := ecc.DecodePoint(structure.SenderRatchetKey, 0)
	if err != nil {
		return nil, err
	}
	var senderRatchetKeyPrivate ecc.ECPrivateKeyable
	if len(structure.SenderRatchetKeyPrivate) == 32 {
		senderRatchetKeyPrivate = ecc.NewDjbECPrivateKey(getArray(structure.SenderRatchetKeyPrivate))
	}
	senderRatchetKeyPair := ecc.NewECKeyPair(senderRatchetKeyPublic, senderRatchetKeyPrivate)

	// Build our message keys from the message key structures.
	messageKeys := make([]*message.Keys, len(structure.MessageKeys))
	for i := range structure.MessageKeys {
		messageKeys[i] = message.NewKeysFromStruct(structure.MessageKeys[i])
	}

	// Build our new chain state.
	chainState := NewChain(
		senderRatchetKeyPair,
		chain.NewKeyFromStruct(structure.ChainKey, kdf.DeriveSecrets),
		messageKeys,
	)

	return chainState, nil
}

// Chain is a structure used inside the SessionState that keeps
// track of an ongoing ratcheting chain for a session.
type Chain struct {
	senderRatchetKeyPair *ecc.ECKeyPair
	chainKey             *chain.Key
	messageKeys          []*message.Keys
}

// SenderRatchetKey returns the sender's EC keypair.
func (c *Chain) SenderRatchetKey() *ecc.ECKeyPair {
	return c.senderRatchetKeyPair
}

// SetSenderRatchetKey will set the chain state with the given EC
// key pair.
func (c *Chain) SetSenderRatchetKey(key *ecc.ECKeyPair) {
	c.senderRatchetKeyPair = key
}

// ChainKey will return the chain key in the chain state.
func (c *Chain) ChainKey() *chain.Key {
	return c.chainKey
}

// SetChainKey will set the chain state's chain key.
func (c *Chain) SetChainKey(key *chain.Key) {
	c.chainKey = key
}

// MessageKeys will return the message keys associated with the
// chain state.
func (c *Chain) MessageKeys() []*message.Keys {
	return c.messageKeys
}

// SetMessageKeys will set the chain state with the given message
// keys.
func (c *Chain) SetMessageKeys(keys []*message.Keys) {
	c.messageKeys = keys
}

// AddMessageKeys will append the chain state with the given
// message keys.
func (c *Chain) AddMessageKeys(keys *message.Keys) {
	c.messageKeys = append(c.messageKeys, keys)
}

// PopFirstMessageKeys will remove the first message key from
// the chain's list of message keys.
func (c *Chain) PopFirstMessageKeys() *message.Keys {
	removed := c.messageKeys[0]
	c.messageKeys = c.messageKeys[1:]

	return removed
}

// structure returns a serializeable structure of the chain state.
func (c *Chain) structure() *pb.SessionStructure_Chain {
	// Alias to ArrayToSlice
	getSlice := bytehelper.ArrayToSlice

	// Convert our message keys into a serializeable structure.
	messageKeys := make([]*pb.SessionStructure_Chain_MessageKey, len(c.messageKeys))
	for i := range c.messageKeys {
		messageKeys[i] = message.NewStructFromKeys(c.messageKeys[i])
	}

	// Convert our sender ratchet key private
	var senderRatchetKeyPrivate []byte
	if c.senderRatchetKeyPair.PrivateKey() != nil {
		senderRatchetKeyPrivate = getSlice(c.senderRatchetKeyPair.PrivateKey().Serialize())
	}

	// Build the chain structure.
	return &pb.SessionStructure_Chain{
		SenderRatchetKey:        c.senderRatchetKeyPair.PublicKey().Serialize(),
		SenderRatchetKeyPrivate: senderRatchetKeyPrivate,
		ChainKey:                chain.NewStructFromKey(c.chainKey),
		MessageKeys:             messageKeys,
	}
}
