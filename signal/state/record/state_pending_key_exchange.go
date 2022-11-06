package record

import (
	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/pb"
	"wa/signal/util/bytehelper"

	"google.golang.org/protobuf/proto"
)

// NewPendingKeyExchange will return a new PendingKeyExchange object.
func NewPendingKeyExchange(sequence uint32, localBaseKeyPair, localRatchetKeyPair *ecc.ECKeyPair,
	localIdentityKeyPair *identity.KeyPair) *PendingKeyExchange {

	return &PendingKeyExchange{
		sequence:             sequence,
		localBaseKeyPair:     localBaseKeyPair,
		localRatchetKeyPair:  localRatchetKeyPair,
		localIdentityKeyPair: localIdentityKeyPair,
	}
}

// NewPendingKeyExchangeFromStruct will return a PendingKeyExchange object from
// the given structure. This is used to get a deserialized pending prekey exchange
// fetched from persistent storage.
func NewPendingKeyExchangeFromStruct(structure *pb.SessionStructure_PendingKeyExchange) *PendingKeyExchange {
	// Return nil if no structure was provided.
	if structure == nil {
		return nil
	}

	// Alias the SliceToArray method.
	getArray := bytehelper.SliceToArray

	// Convert the bytes in the given structure to ECC objects.
	localBaseKeyPair := ecc.NewECKeyPair(
		ecc.NewDjbECPublicKey(structure.LocalBaseKey),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalBaseKeyPrivate)),
	)
	localRatchetKeyPair := ecc.NewECKeyPair(
		ecc.NewDjbECPublicKey(structure.LocalRatchetKey),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalRatchetKeyPrivate)),
	)
	localIdentityKeyPair := identity.NewKeyPair(
		identity.NewKey(ecc.NewDjbECPublicKey(structure.LocalIdentityKey)),
		ecc.NewDjbECPrivateKey(getArray(structure.LocalIdentityKeyPrivate)),
	)

	// Return the PendingKeyExchange with the deserialized keys.
	return &PendingKeyExchange{
		sequence:             structure.GetSequence(),
		localBaseKeyPair:     localBaseKeyPair,
		localRatchetKeyPair:  localRatchetKeyPair,
		localIdentityKeyPair: localIdentityKeyPair,
	}
}

// PendingKeyExchange is a structure for storing a pending
// key exchange for a session state.
type PendingKeyExchange struct {
	sequence             uint32
	localBaseKeyPair     *ecc.ECKeyPair
	localRatchetKeyPair  *ecc.ECKeyPair
	localIdentityKeyPair *identity.KeyPair
}

// structure will return a serializable structure of a pending key exchange
// so it can be persistently stored.
func (p *PendingKeyExchange) structure() *pb.SessionStructure_PendingKeyExchange {
	getSlice := bytehelper.ArrayToSlice
	return &pb.SessionStructure_PendingKeyExchange{
		Sequence:                proto.Uint32(p.sequence),
		LocalBaseKey:            getSlice(p.localBaseKeyPair.PublicKey().PublicKey()),
		LocalBaseKeyPrivate:     getSlice(p.localBaseKeyPair.PrivateKey().Serialize()),
		LocalRatchetKey:         getSlice(p.localRatchetKeyPair.PublicKey().PublicKey()),
		LocalRatchetKeyPrivate:  getSlice(p.localRatchetKeyPair.PrivateKey().Serialize()),
		LocalIdentityKey:        getSlice(p.localIdentityKeyPair.PublicKey().PublicKey().PublicKey()),
		LocalIdentityKeyPrivate: getSlice(p.localIdentityKeyPair.PrivateKey().Serialize()),
	}
}
