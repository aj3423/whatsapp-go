package protocol

import (
	"errors"
	"strconv"

	"wa/signal/ecc"
	"wa/signal/pb"

	"github.com/golang/protobuf/proto"
)

func NewSenderKeyDistributionMessageFromBytes(
	data []byte,
) (*SenderKeyDistributionMessage, error) {
	if len(data) < 2 {
		return nil, errors.New(`skdm too short`)
	}

	p := &pb.SenderKeyDistributionMessage{}
	e := proto.Unmarshal(data[1:], p)
	if e != nil {
		return nil, e
	}

	ver, e := strconv.Atoi(string(data[0:1]))
	if e != nil {
		return nil, e
	}
	stru := &SenderKeyDistributionMessage{
		P:       p,
		version: uint32(ver),
	}

	e = verify_sender_key_distribution_message(stru)
	if e != nil {
		return nil, e
	}
	return stru, nil
}

func verify_sender_key_distribution_message(
	structure *SenderKeyDistributionMessage,
) error {

	// Throw an error if the given message structure is an unsupported version.
	if structure.version <= UnsupportedVersion {
		err := "Legacy message: " + strconv.Itoa(int(structure.version))
		return errors.New(err)
	}

	// Throw an error if the given message structure is a future version.
	if structure.version > CurrentVersion {
		err := "Unknown version: " + strconv.Itoa(int(structure.version))
		return errors.New(err)
	}

	// Throw an error if the structure is missing critical fields.
	if len(structure.P.GetSigningKey()) == 0 || len(structure.P.GetChainKey()) == 0 {
		err := "Incomplete message."
		return errors.New(err)
	}

	// Get the signing key object from bytes.
	_, err := ecc.DecodePoint(structure.P.GetSigningKey(), 0)
	if err != nil {
		return err
	}

	return nil
}

// NewSenderKeyDistributionMessage returns a Signal Ciphertext message.
func NewSenderKeyDistributionMessage(
	id uint32, iteration uint32,
	chainKey []byte, signatureKey ecc.ECPublicKeyable,
) *SenderKeyDistributionMessage {

	k := signatureKey.PublicKey()
	return &SenderKeyDistributionMessage{
		version: CurrentVersion,
		P: &pb.SenderKeyDistributionMessage{
			Id:         &id,
			Iteration:  &iteration,
			ChainKey:   chainKey,
			SigningKey: k[:],
		},
	}
}

type SenderKeyDistributionMessage struct {
	version uint32
	P       *pb.SenderKeyDistributionMessage
}

func (p *SenderKeyDistributionMessage) ID() uint32 {
	return p.P.GetId()
}

func (p *SenderKeyDistributionMessage) Iteration() uint32 {
	return p.P.GetIteration()
}

func (p *SenderKeyDistributionMessage) ChainKey() []byte {
	return p.P.GetChainKey()
}

func (p *SenderKeyDistributionMessage) SignatureKey() ecc.ECPublicKeyable {
	signingKey, err := ecc.DecodePoint(p.P.GetSigningKey(), 0)
	if err != nil {
		return nil
	}

	return signingKey
}

func (p *SenderKeyDistributionMessage) Serialize() []byte {
	// make a copy, the SingingKey should begin with '05'
	p2 := &pb.SenderKeyDistributionMessage{
		Id:         p.P.Id,
		Iteration:  p.P.Iteration,
		ChainKey:   p.P.ChainKey,
		SigningKey: append([]byte{ecc.DjbType}, p.P.SigningKey...),
	}

	// version + proto
	bs, e := proto.Marshal(p2)
	if e != nil {
		return nil
	}
	ver := []byte(strconv.Itoa(int(p.version)))
	ret := append(ver, bs...)
	return ret
}

// Type will return the message's type.
func (p *SenderKeyDistributionMessage) Type() uint32 {
	return SENDERKEY_DISTRIBUTION_TYPE
}
