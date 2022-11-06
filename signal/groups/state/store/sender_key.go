package store

import (
	"wa/signal/groups/state/record"
	"wa/signal/protocol"
)

type SenderKey interface {
	StoreSenderKey(senderKeyName *protocol.SenderKeyName, keyRecord *record.SenderKey) error
	LoadSenderKey(senderKeyName *protocol.SenderKeyName) (*record.SenderKey, error)
}
