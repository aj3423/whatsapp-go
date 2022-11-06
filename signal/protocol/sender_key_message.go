package protocol

import (
	"errors"
	"strconv"

	"wa/signal/ecc"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"

	"github.com/golang/protobuf/proto"
)

func NewSenderKeyMessageFromBytes(data []byte) (*SenderKeyMessage, error) {
	if len(data) < 65 { // 1byte ver,  64 byte signature
		return nil, errors.New(`skmsg too short`)
	}

	pb_body := data[1 : len(data)-64]
	p := &pb.SenderKeyMessage{}
	e := proto.Unmarshal(pb_body, p)
	if e != nil {
		return nil, e
	}

	ver, e := strconv.Atoi(string(data[0:1]))
	if e != nil {
		return nil, e
	}

	stru := &SenderKeyMessage{
		P:         p,
		version:   uint32(ver),
		signature: data[len(data)-64:],
	}

	e = verify_sender_key_message(stru)
	if e != nil {
		return nil, e
	}

	return stru, nil
}

func verify_sender_key_message(msg *SenderKeyMessage) error {
	// Throw an error if the given message msg is an unsupported version.
	if msg.version <= UnsupportedVersion {
		err := "Legacy message: " + strconv.Itoa(int(msg.version))
		return errors.New(err)
	}

	// Throw an error if the given message msg is a future version.
	if msg.version > CurrentVersion {
		err := "Unknown version: " + strconv.Itoa(int(msg.version))
		return errors.New(err)
	}

	if len(msg.P.GetCiphertext()) == 0 {
		err := "Incomplete message."
		return errors.New(err)
	}

	return nil
}

// NewSenderKeyMessage returns a SenderKeyMessage.
func NewSenderKeyMessage(
	keyID uint32, iteration uint32, ciphertext []byte,
	signatureKey ecc.ECPrivateKeyable,
) (*SenderKeyMessage, error) {

	// Ensure we have a valid signature key
	if signatureKey == nil {
		return nil, errors.New("Signature is nil")
	}

	// Build our SenderKeyMessage.
	msg := &SenderKeyMessage{
		P: &pb.SenderKeyMessage{
			Id:         &keyID,
			Iteration:  &iteration,
			Ciphertext: ciphertext,
		},
		version: CurrentVersion,
	}

	// Sign the serialized message and include it in the message. This will be included
	// in the signed serialized version of the message.
	signature, e := ecc.CalculateSignature(signatureKey, msg.Serialize())
	if e != nil {
		return nil, e
	}
	msg.signature = bytehelper.ArrayToSlice64(signature)

	return msg, nil
}

// SenderKeyMessage is a structure for messages using senderkey groups.
type SenderKeyMessage struct {
	P         *pb.SenderKeyMessage
	version   uint32
	signature []byte
}

// KeyID returns the SenderKeyMessage key ID.
func (p *SenderKeyMessage) KeyID() uint32 {
	return p.P.GetId()
}

// Iteration returns the SenderKeyMessage iteration.
func (p *SenderKeyMessage) Iteration() uint32 {
	return p.P.GetIteration()
}

// Ciphertext returns the SenderKeyMessage encrypted ciphertext.
func (p *SenderKeyMessage) Ciphertext() []byte {
	return p.P.GetCiphertext()
}

// Version returns the Signal message version of the message.
func (p *SenderKeyMessage) Version() uint32 {
	return p.version
}

func (p *SenderKeyMessage) Serialize() []byte {
	bs, e := proto.Marshal(p.P)
	if e != nil {
		return nil
	}
	ver := []byte(strconv.Itoa(int(p.version)))
	ret := append(ver, bs...)
	return ret
	/*
		structure := &SenderKeyMessageStructure{
			ID:         p.keyID,
			Iteration:  p.iteration,
			CipherText: p.ciphertext,

			Version:    p.version,
		}
		return p.serializer.Serialize(structure)
	*/
}

func (p *SenderKeyMessage) SignedSerialize() []byte {
	bs, e := proto.Marshal(p.P)
	if e != nil {
		return nil
	}

	ver := []byte(strconv.Itoa(int(p.version)))
	ret := append(ver, bs...)
	ret = append(ret, p.signature...)
	return ret

	/*
		structure := &SenderKeyMessageStructure{
			ID:         p.keyID,
			Iteration:  p.iteration,
			CipherText: p.ciphertext,
			Version:    p.version,
			Signature:  p.signature,
		}
		return p.serializer.Serialize(structure)
	*/
}

// Signature returns the SenderKeyMessage signature
func (p *SenderKeyMessage) Signature() [64]byte {
	return bytehelper.SliceToArray64(p.signature)
}

// Type returns the sender key type.
func (p *SenderKeyMessage) Type() uint32 {
	return SENDERKEY_TYPE
}
