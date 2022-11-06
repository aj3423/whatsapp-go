package record

import (
	"wa/signal/ecc"
	"wa/signal/pb"
	"wa/signal/util/optional"

	"google.golang.org/protobuf/proto"
)

// NewPendingPreKey will return a new pending pre key object.
func NewPendingPreKey(preKeyID *optional.Uint32, signedPreKeyID uint32,
	baseKey ecc.ECPublicKeyable) *PendingPreKey {

	return &PendingPreKey{
		preKeyID:       preKeyID,
		signedPreKeyID: signedPreKeyID,
		baseKey:        baseKey,
	}
}

// NewPendingPreKeyFromStruct will return a new pending prekey object from the
// given structure.
func NewPendingPreKeyFromStruct(preKey *pb.SessionStructure_PendingPreKey) (*PendingPreKey, error) {
	baseKey, err := ecc.DecodePoint(preKey.BaseKey, 0)
	if err != nil {
		return nil, err
	}

	id := optional.NewOptionalUint32(preKey.GetPreKeyId())
	if preKey.PreKeyId == nil {
		id.IsEmpty = true
	}

	pendingPreKey := NewPendingPreKey(
		id,
		uint32(preKey.GetSignedPreKeyId()),
		baseKey,
	)

	return pendingPreKey, nil
}

// PendingPreKey is a structure for pending pre keys
// for a session state.
type PendingPreKey struct {
	preKeyID       *optional.Uint32
	signedPreKeyID uint32
	baseKey        ecc.ECPublicKeyable
}

// structure will return a serializeable structure of the pending prekey.
func (p *PendingPreKey) structure() *pb.SessionStructure_PendingPreKey {
	if p != nil {
		ret := &pb.SessionStructure_PendingPreKey{
			SignedPreKeyId: proto.Int32(int32(p.signedPreKeyID)),
			BaseKey:        p.baseKey.Serialize(),
		}
		if !p.preKeyID.IsEmpty {
			ret.PreKeyId = proto.Uint32(p.preKeyID.Value)
		}
		return ret
	}
	return nil
}
