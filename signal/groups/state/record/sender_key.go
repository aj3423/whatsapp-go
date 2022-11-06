package record

import (
	"errors"

	"wa/signal/ecc"
	"wa/signal/pb"

	"github.com/golang/protobuf/proto"
)

func NewSenderKeyFromBytes(data []byte) (*SenderKey, error) {
	p := &pb.SenderKeyRecordStructure{}
	e := proto.Unmarshal(data, p)
	if e != nil {
		return nil, e
	}

	// Build our sender key states from structure.
	senderKeyStates := make([]*SenderKeyState, len(p.SenderKeyStates))
	for i := range p.SenderKeyStates {
		var err error
		senderKeyStates[i], err = NewSenderKeyStateFromStructure(p.SenderKeyStates[i])
		if err != nil {
			return nil, err
		}
	}

	// Build and return our session.
	senderKey := &SenderKey{
		senderKeyStates: senderKeyStates,
	}

	return senderKey, nil

}

// NewSenderKey record returns a new sender key record that can
// be stored in a SenderKeyStore.
func NewSenderKey() *SenderKey {
	return &SenderKey{
		senderKeyStates: []*SenderKeyState{},
	}
}

// SenderKey record is a structure for storing pre keys inside
// a SenderKeyStore.
type SenderKey struct {
	senderKeyStates []*SenderKeyState
}

// SenderKeyState will return the first sender key state in the record's
// list of sender key states.
func (k *SenderKey) SenderKeyState() (*SenderKeyState, error) {
	if len(k.senderKeyStates) > 0 {
		return k.senderKeyStates[0], nil
	}
	return nil, errors.New("No Sender Keys")
}

// GetSenderKeyStateByID will return the sender key state with the given
// key id.
func (k *SenderKey) GetSenderKeyStateByID(keyID uint32) (*SenderKeyState, error) {
	for i := 0; i < len(k.senderKeyStates); i++ {
		if k.senderKeyStates[i].KeyID() == keyID {
			return k.senderKeyStates[i], nil
		}
	}

	return nil, errors.New("No sender key for for ID")
}

// IsEmpty will return false if there is more than one state in this
// senderkey record.
func (k *SenderKey) IsEmpty() bool {
	return len(k.senderKeyStates) == 0
}

// AddSenderKeyState will add a new state to this senderkey record with the given
// id, iteration, chainkey, and signature key.
func (k *SenderKey) AddSenderKeyState(
	id uint32, iteration uint32,
	chainKey []byte, signatureKey ecc.ECPublicKeyable,
) {

	newState := NewSenderKeyStateFromPublicKey(id, iteration, chainKey, signatureKey)
	k.senderKeyStates = append(k.senderKeyStates, newState)

	if len(k.senderKeyStates) > maxMessageKeys {
		k.senderKeyStates = k.senderKeyStates[1:]
	}
}

// SetSenderKeyState will  replace the current senderkey states with the given
// senderkey state.
func (k *SenderKey) SetSenderKeyState(
	id uint32, iteration uint32,
	chainKey []byte, signatureKey *ecc.ECKeyPair,
) {

	newState := NewSenderKeyState(id, iteration, chainKey, signatureKey)
	k.senderKeyStates = make([]*SenderKeyState, 0, maxMessageKeys/2)
	k.senderKeyStates = append(k.senderKeyStates, newState)
}

// Serialize will return the record as serialized bytes so it can be
// persistently stored.
func (k *SenderKey) Serialize() ([]byte, error) {
	p := k.structure()
	return proto.Marshal(p)
}

// structure will return a simple serializable record structure.
// This is used for serialization to persistently
// store a session record.
func (k *SenderKey) structure() *pb.SenderKeyRecordStructure {
	p := &pb.SenderKeyRecordStructure{}
	for _, sta := range k.senderKeyStates {
		p.SenderKeyStates = append(p.SenderKeyStates, sta.structure())
	}
	return p
}
