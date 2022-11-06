package protocol

import (
	"errors"
	"strconv"

	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/pb"
	"wa/signal/util/optional"

	"github.com/golang/protobuf/proto"
)

// NewPreKeySignalMessageFromBytes will return a Signal Ciphertext message from the given
// bytes using the given serializer.
func NewPreKeySignalMessageFromBytes(data []byte) (*PreKeySignalMessage, error) {
	if len(data) < 2 {
		return nil, errors.New(`pkmsg too short`)
	}

	p := &pb.PreKeySignalMessage{}
	e := proto.Unmarshal(data[1:], p)
	if e != nil {
		return nil, e
	}

	ver, e := strconv.Atoi(string(data[0:1]))
	if e != nil {
		return nil, e
	}
	stru := &PreKeySignalMessageStructure{
		P:       p,
		Version: ver,
	}

	return NewPreKeySignalMessageFromStruct(stru)
}

// NewPreKeySignalMessageFromStruct will return a new PreKeySignalMessage from the given
// PreKeySignalMessageStructure.
func NewPreKeySignalMessageFromStruct(
	structure *PreKeySignalMessageStructure) (*PreKeySignalMessage, error) {

	// Throw an error if the given message structure is an unsupported version.
	if structure.Version <= UnsupportedVersion {
		err := "Legacy message: " + strconv.Itoa(structure.Version)
		return nil, errors.New(err)
	}

	// Throw an error if the given message structure is a future version.
	if structure.Version > CurrentVersion {
		err := "Unknown version: " + strconv.Itoa(structure.Version)
		return nil, errors.New(err)
	}

	// Throw an error if the structure is missing critical fields.
	if structure.P.BaseKey == nil || structure.P.IdentityKey == nil || structure.P.Message == nil {
		err := "Incomplete message."
		return nil, errors.New(err)
	}

	// Create the signal message object from the structure.
	preKeyWhisperMessage := &PreKeySignalMessage{structure: *structure}

	// Generate the base ECC key from bytes.
	var err error
	preKeyWhisperMessage.baseKey, err = ecc.DecodePoint(structure.P.GetBaseKey(), 0)
	if err != nil {
		return nil, err
	}

	// Generate the identity key from bytes
	var identityKey ecc.ECPublicKeyable
	identityKey, err = ecc.DecodePoint(structure.P.GetIdentityKey(), 0)
	if err != nil {
		return nil, err
	}
	preKeyWhisperMessage.identityKey = identity.NewKey(identityKey)

	// Generate the SignalMessage object from bytes.
	preKeyWhisperMessage.message, err = NewSignalMessageFromBytes(structure.P.GetMessage())
	if err != nil {
		return nil, err
	}

	return preKeyWhisperMessage, nil
}

func NewPreKeySignalMessage(
	version int, registrationID uint32, preKeyID *optional.Uint32, signedPreKeyID uint32,
	baseKey ecc.ECPublicKeyable, identityKey *identity.Key, message *SignalMessage,
) (*PreKeySignalMessage, error) {
	p := &pb.PreKeySignalMessage{
		RegistrationId: proto.Uint32(registrationID),
		SignedPreKeyId: proto.Uint32(signedPreKeyID),
		BaseKey:        baseKey.Serialize(),
		IdentityKey:    identityKey.PublicKey().Serialize(),
		Message:        message.Serialize(),
	}
	if !preKeyID.IsEmpty {
		p.PreKeyId = &preKeyID.Value
	}

	structure := &PreKeySignalMessageStructure{
		Version: version,
		P:       p,
	}
	return NewPreKeySignalMessageFromStruct(structure)
}

type PreKeySignalMessageStructure struct {
	P *pb.PreKeySignalMessage

	Version int
}

type PreKeySignalMessage struct {
	structure   PreKeySignalMessageStructure
	baseKey     ecc.ECPublicKeyable
	identityKey *identity.Key
	message     *SignalMessage
}

func (p *PreKeySignalMessage) MessageVersion() int {
	return p.structure.Version
}

func (p *PreKeySignalMessage) IdentityKey() *identity.Key {
	return p.identityKey
}

func (p *PreKeySignalMessage) RegistrationID() uint32 {
	return p.structure.P.GetRegistrationId()
}

func (p *PreKeySignalMessage) PreKeyID() *optional.Uint32 {
	if p.structure.P.PreKeyId == nil {
		return nil
		//return optional.NewEmptyUint32()
	} else {
		return optional.NewOptionalUint32(p.structure.P.GetPreKeyId())
	}
}

func (p *PreKeySignalMessage) SignedPreKeyID() uint32 {
	return p.structure.P.GetSignedPreKeyId()
}

func (p *PreKeySignalMessage) BaseKey() ecc.ECPublicKeyable {
	return p.baseKey
}

func (p *PreKeySignalMessage) WhisperMessage() *SignalMessage {
	return p.message
}

func (p *PreKeySignalMessage) Serialize() []byte {
	ret, e := proto.Marshal(p.structure.P)
	if e != nil {
		return nil
	}
	ver := []byte(strconv.Itoa(p.structure.Version))
	return append(ver, ret...)
}

func (p *PreKeySignalMessage) Type() uint32 {
	return PREKEY_TYPE
}
