// Package groups is responsible for setting up group SenderKey encrypted sessions.
// Once a session has been established, GroupCipher can be used to encrypt/decrypt
// messages in that session.
//
// The built sessions are unidirectional: they can be used either for sending or
// for receiving, but not both. Sessions are constructed per (groupId + senderId +
// deviceId) tuple. Remote logical users are identified by their senderId, and each
// logical recipientId can have multiple physical devices.
package groups

import (
	"wa/signal/groups/state/record"
	"wa/signal/groups/state/store"
	"wa/signal/protocol"
	"wa/signal/util/keyhelper"
)

// NewGroupSessionBuilder will return a new group session builder.
func NewGroupSessionBuilder(
	senderKeyStore store.SenderKey,
) *SessionBuilder {

	return &SessionBuilder{
		senderKeyStore: senderKeyStore,
	}
}

// SessionBuilder is a structure for building group sessions.
type SessionBuilder struct {
	senderKeyStore store.SenderKey
}

// Process will process an incoming group message and set up the corresponding
// session for it.
func (b *SessionBuilder) Process(senderKeyName *protocol.SenderKeyName,
	msg *protocol.SenderKeyDistributionMessage) error {

	senderKeyRecord, e := b.senderKeyStore.LoadSenderKey(senderKeyName)
	if e != nil {
		return e
	}
	if senderKeyRecord == nil {
		senderKeyRecord = record.NewSenderKey()
	}
	senderKeyRecord.AddSenderKeyState(msg.ID(), msg.Iteration(), msg.ChainKey(), msg.SignatureKey())
	return b.senderKeyStore.StoreSenderKey(senderKeyName, senderKeyRecord)
}

// Create will create a new group session for the given name.
func (b *SessionBuilder) Create(senderKeyName *protocol.SenderKeyName) (*protocol.SenderKeyDistributionMessage, error) {
	// Load the senderkey by name
	senderKeyRecord, e := b.senderKeyStore.LoadSenderKey(senderKeyName)
	if e != nil {
		return nil, e
	}

	// If the record is empty, generate new keys.
	if senderKeyRecord == nil || senderKeyRecord.IsEmpty() {
		senderKeyRecord = record.NewSenderKey()
		signingKey, err := keyhelper.GenerateSenderSigningKey()
		if err != nil {
			return nil, err
		}
		senderKeyRecord.SetSenderKeyState(
			keyhelper.GenerateSenderKeyID(), 0,
			keyhelper.GenerateSenderKey(),
			signingKey,
		)
		b.senderKeyStore.StoreSenderKey(senderKeyName, senderKeyRecord)
	}

	// Get the senderkey state.
	state, err := senderKeyRecord.SenderKeyState()
	if err != nil {
		return nil, err
	}

	// Create the group message to return.
	senderKeyDistributionMessage := protocol.NewSenderKeyDistributionMessage(
		state.KeyID(),
		state.SenderChainKey().Iteration(),
		state.SenderChainKey().Seed(),
		state.SigningKey().PublicKey(),
	)

	return senderKeyDistributionMessage, nil
}
