package protocol

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strconv"

	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"

	"github.com/golang/protobuf/proto"
)

const MacLength int = 8

// data:
// 33 ........... Mac-8-byte
func NewSignalMessageFromBytes(data []byte) (*SignalMessage, error) {
	if len(data) < 9 {
		return nil, errors.New(`msg too short`)
	}
	pb_body := data[1 : len(data)-MacLength]

	p := &pb.SignalMessage{}
	e := proto.Unmarshal(pb_body, p)
	if e != nil {
		return nil, e
	}

	ver, e := strconv.Atoi(string(data[0:1]))
	if e != nil {
		return nil, e
	}

	stru := &SignalMessageStructure{
		P:       p,
		Version: ver,
		Mac:     data[len(data)-MacLength:], // last 8
	}

	return NewSignalMessageFromStruct(stru)
}

// NewSignalMessageFromStruct returns a Signal Ciphertext message from the
// given serializable structure.
func NewSignalMessageFromStruct(structure *SignalMessageStructure) (*SignalMessage, error) {
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
	if structure.P.Ciphertext == nil || structure.P.RatchetKey == nil {
		err := "Incomplete message."
		return nil, errors.New(err)
	}

	// Create the signal message object from the structure.
	whisperMessage := &SignalMessage{structure: *structure}

	// Generate the ECC key from bytes.
	var err error
	whisperMessage.senderRatchetKey, err = ecc.DecodePoint(structure.P.RatchetKey, 0)
	if err != nil {
		return nil, err
	}

	return whisperMessage, nil
}

func NewSignalMessage(messageVersion int, counter, previousCounter uint32, macKey []byte,
	senderRatchetKey ecc.ECPublicKeyable, ciphertext []byte, senderIdentityKey,
	receiverIdentityKey *identity.Key) (*SignalMessage, error) {

	version := []byte(strconv.Itoa(messageVersion))
	// Build the signal message structure with the given data.
	structure := &SignalMessageStructure{
		P: &pb.SignalMessage{
			Counter:         proto.Uint32(counter),
			PreviousCounter: proto.Uint32(previousCounter),
			RatchetKey:      senderRatchetKey.Serialize(),
			Ciphertext:      ciphertext,
		},
	}
	p_data, e := proto.Marshal(structure.P)
	if e != nil {
		return nil, e
	}

	serialized := append(version, p_data...)
	// Get the message authentication code from the serialized structure.
	mac, err := getMac(
		messageVersion, senderIdentityKey, receiverIdentityKey,
		macKey, serialized,
	)
	if err != nil {
		return nil, err
	}
	structure.Mac = mac
	structure.Version = messageVersion

	// Generate a SignalMessage with the structure.
	whisperMessage, err := NewSignalMessageFromStruct(structure)
	if err != nil {
		return nil, err
	}

	return whisperMessage, nil
}

// SignalMessageStructure is a serializeable structure of a signal message
// object.
type SignalMessageStructure struct {
	P       *pb.SignalMessage
	Version int
	Mac     []byte
}

// SignalMessage is a cipher message that contains a message encrypted
// with the Signal protocol.
type SignalMessage struct {
	structure        SignalMessageStructure
	senderRatchetKey ecc.ECPublicKeyable
}

// SenderRatchetKey returns the SignalMessage's sender ratchet key. This
// key is used for ratcheting the chain forward to negotiate a new shared
// secret that cannot be derived from previous chains.
func (s *SignalMessage) SenderRatchetKey() ecc.ECPublicKeyable {
	return s.senderRatchetKey
}

// MessageVersion returns the message version this SignalMessage supports.
func (s *SignalMessage) MessageVersion() int {
	return s.structure.Version
}

// Counter will return the SignalMessage counter.
func (s *SignalMessage) Counter() uint32 {
	return s.structure.P.GetCounter()
}

// Body will return the SignalMessage's ciphertext in bytes.
func (s *SignalMessage) Body() []byte {
	return s.structure.P.GetCiphertext()
}

// VerifyMac will return an error if the message's message authentication code
// is invalid. This should be used on SignalMessages that have been constructed
// from a sent message.
func (s *SignalMessage) VerifyMac(messageVersion int, senderIdentityKey,
	receiverIdentityKey *identity.Key, macKey []byte) error {

	// Create a copy of the message without the mac. We'll use this to calculate
	// the message authentication code.
	structure := s.structure
	signalMessage, err := NewSignalMessageFromStruct(&structure)
	if err != nil {
		return err
	}
	signalMessage.structure.Mac = nil
	signalMessage.structure.Version = 0
	version := []byte(strconv.Itoa(s.MessageVersion()))
	serialized := append(version, signalMessage.Serialize()...)

	// Calculate the message authentication code from the serialized structure.
	ourMac, err := getMac(
		messageVersion,
		senderIdentityKey,
		receiverIdentityKey,
		macKey,
		serialized,
	)
	if err != nil {
		return err
	}

	// Get the message authentication code that was sent to us as part of
	// the signal message structure.
	theirMac := s.structure.Mac

	// Return an error if our calculated mac doesn't match the mac sent to us.
	if !hmac.Equal(ourMac, theirMac) {
		return errors.New("Bad Mac!")
	}

	return nil
}

func (s *SignalMessage) Serialize() []byte {
	bs, e := proto.Marshal(s.structure.P)
	if e != nil {
		return nil
	}

	// VerifyMac will set Version/Mac to 0
	// only need the protobuf data
	if s.structure.Version == 0 {
		return bs
	}

	ver := []byte(strconv.Itoa(s.structure.Version))
	ret := append(ver, bs...)
	ret = append(ret, s.structure.Mac...)
	return ret
}

// Structure will return a serializeable structure of the Signal Message.
func (s *SignalMessage) Structure() *SignalMessageStructure {
	structure := s.structure
	return &structure
}

// Type will return the type of Signal Message this is.
func (s *SignalMessage) Type() uint32 {
	return WHISPER_TYPE
}

// getMac will calculate the mac using the given message version, identity
// keys, macKey and SignalMessageStructure. The MAC key is a private key held
// by both parties that is concatenated with the message and hashed.
func getMac(messageVersion int, senderIdentityKey, receiverIdentityKey *identity.Key,
	macKey, serialized []byte) ([]byte, error) {

	mac := hmac.New(sha256.New, macKey[:])

	if messageVersion >= 3 {
		mac.Write(senderIdentityKey.PublicKey().Serialize())
		mac.Write(receiverIdentityKey.PublicKey().Serialize())
	}

	mac.Write(serialized)

	fullMac := mac.Sum(nil)

	return bytehelper.Trim(fullMac, MacLength), nil
}
