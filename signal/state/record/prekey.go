package record

import (
	"wa/signal/ecc"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"
	"wa/signal/util/optional"

	"google.golang.org/protobuf/proto"
)

// NewPreKeyFromBytes will return a prekey record from the given bytes using the given serializer.
func NewPreKeyFromBytes(serialized []byte) (*PreKey, error) {
	p := pb.PreKeyRecordStructure{}
	e := proto.Unmarshal(serialized, &p)
	if e != nil {
		return nil, e
	}
	return NewPreKeyFromStruct(&p)
}

func NewPreKeyFromStruct(structure *pb.PreKeyRecordStructure) (*PreKey, error) {
	preKey := &PreKey{
		structure: structure,
	}

	// Generate the ECC key from bytes.
	publicKey := ecc.NewDjbECPublicKey(structure.PublicKey)
	privateKey := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(structure.PrivateKey))
	keyPair := ecc.NewECKeyPair(publicKey, privateKey)
	preKey.keyPair = keyPair

	return preKey, nil
}

// NewPreKey record returns a new pre key record that can
// be stored in a PreKeyStore.
func NewPreKey(id uint32, keyPair *ecc.ECKeyPair) *PreKey {
	return &PreKey{
		structure: &pb.PreKeyRecordStructure{
			Id:         proto.Uint32(id),
			PublicKey:  keyPair.PublicKey().Serialize(),
			PrivateKey: bytehelper.ArrayToSlice(keyPair.PrivateKey().Serialize()),
		},
		keyPair: keyPair,
	}
}

// PreKey record is a structure for storing pre keys inside
// a PreKeyStore.
type PreKey struct {
	structure *pb.PreKeyRecordStructure
	keyPair   *ecc.ECKeyPair
}

// ID returns the pre key record's id.
func (p *PreKey) ID() *optional.Uint32 {
	return optional.NewOptionalUint32(p.structure.GetId())
}

// KeyPair returns the pre key record's key pair.
func (p *PreKey) KeyPair() *ecc.ECKeyPair {
	return p.keyPair
}

// Serialize uses the PreKey serializer to return the PreKey
// as serialized bytes.
func (p *PreKey) Serialize() ([]byte, error) {
	return proto.Marshal(p.structure)
}
