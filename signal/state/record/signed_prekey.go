package record

import (
	"wa/signal/ecc"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"

	"github.com/golang/protobuf/proto"
)

func NewSignedPreKeyFromBytes(serialized []byte) (*SignedPreKey, error) {
	p := &pb.SignedPreKeyRecordStructure{}
	e := proto.Unmarshal(serialized, p)
	if e != nil {
		return nil, e
	}
	return NewSignedPreKeyFromStruct(p)
}

// NewSignedPreKeyFromStruct returns a SignedPreKey record using the given
// serializable structure.
func NewSignedPreKeyFromStruct(structure *pb.SignedPreKeyRecordStructure) (*SignedPreKey, error) {

	// Create the signed prekey record from the structure.
	signedPreKey := &SignedPreKey{
		structure: structure,
		signature: bytehelper.SliceToArray64(structure.Signature),
	}

	// Generate the ECC key from bytes.
	publicKey := ecc.NewDjbECPublicKey(structure.PublicKey)
	privateKey := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.PrivateKey))
	keyPair := ecc.NewECKeyPair(publicKey, privateKey)
	signedPreKey.keyPair = keyPair

	return signedPreKey, nil
}

// NewSignedPreKey record creates a new signed pre key record
// with the given properties.
func NewSignedPreKey(
	id uint32,
	timestamp int64,
	keyPair *ecc.ECKeyPair,
	sig [64]byte,
) *SignedPreKey {

	return &SignedPreKey{
		structure: &pb.SignedPreKeyRecordStructure{
			Id:         proto.Uint32(id),
			Timestamp:  proto.Uint64(uint64(timestamp)),
			PublicKey:  keyPair.PublicKey().Serialize(),
			PrivateKey: bytehelper.ArrayToSlice(keyPair.PrivateKey().Serialize()),
			Signature:  bytehelper.ArrayToSlice64(sig),
		},
		keyPair:   keyPair,
		signature: sig,
	}
}

// SignedPreKey record is a structure for storing a signed
// pre key in a SignedPreKey store.
type SignedPreKey struct {
	structure *pb.SignedPreKeyRecordStructure
	keyPair   *ecc.ECKeyPair
	signature [64]byte
}

// ID returns the record's id.
func (s *SignedPreKey) ID() uint32 {
	return s.structure.GetId()
}

// Timestamp returns the record's timestamp
func (s *SignedPreKey) Timestamp() int64 {
	return int64(s.structure.GetTimestamp())
}

// KeyPair returns the signed pre key record's key pair.
func (s *SignedPreKey) KeyPair() *ecc.ECKeyPair {
	return s.keyPair
}

// Signature returns the record's signed prekey signature.
func (s *SignedPreKey) Signature() [64]byte {
	return s.signature
}

// Serialize uses the SignedPreKey serializer to return the SignedPreKey
// as serialized bytes.
func (s *SignedPreKey) Serialize() []byte {
	bs, e := proto.Marshal(s.structure)
	if e != nil {
		return nil
	}
	return bs
}
