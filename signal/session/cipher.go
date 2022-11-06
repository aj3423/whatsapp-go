package session

import (
	"errors"
	"strconv"

	"wa/signal/cipher"
	"wa/signal/ecc"
	"wa/signal/keys/chain"
	"wa/signal/keys/message"
	"wa/signal/protocol"
	"wa/signal/state/record"
	"wa/signal/state/store"
	"wa/signal/util/bytehelper"
)

const maxFutureMessages = 5000

// NewCipher constructs a session cipher for encrypt/decrypt operations on a
// session. In order to use the session cipher, a session must have already
// been created and stored using session.Builder.
func NewCipher(
	builder *Builder,
	remoteAddress *protocol.SignalAddress,
) *Cipher {
	return &Cipher{
		sessionStore:     builder.sessionStore,
		preKeyStore:      builder.preKeyStore,
		remoteAddress:    remoteAddress,
		builder:          builder,
		identityKeyStore: builder.identityKeyStore,
	}
}

func NewCipherFromSession(
	remoteAddress *protocol.SignalAddress,
	sessionStore store.Session, preKeyStore store.PreKey, identityKeyStore store.IdentityKey,
) *Cipher {

	return &Cipher{
		sessionStore:     sessionStore,
		preKeyStore:      preKeyStore,
		remoteAddress:    remoteAddress,
		identityKeyStore: identityKeyStore,
	}
}

// Cipher is the main entry point for Signal Protocol encrypt/decrypt operations.
// Once a session has been established with session.Builder, this can be used for
// all encrypt/decrypt operations within that session.
type Cipher struct {
	sessionStore     store.Session
	preKeyStore      store.PreKey
	remoteAddress    *protocol.SignalAddress
	builder          *Builder
	identityKeyStore store.IdentityKey
}

// Encrypt will take the given message in bytes and return an object that follows
// the CiphertextMessage interface.
func (d *Cipher) Encrypt(plaintext []byte) (protocol.CiphertextMessage, error) {
	sessionRecord, e := d.sessionStore.LoadSession(d.remoteAddress)
	if e != nil {
		return nil, e
	}
	sessionState := sessionRecord.SessionState()
	chainKey := sessionState.SenderChainKey()
	messageKeys := chainKey.MessageKeys()
	senderEphemeral := sessionState.SenderRatchetKey()
	previousCounter := sessionState.PreviousCounter()
	sessionVersion := sessionState.Version()

	ciphertextBody, err := encrypt(messageKeys, plaintext)
	if err != nil {
		return nil, err
	}

	var ciphertextMessage protocol.CiphertextMessage
	ciphertextMessage, err = protocol.NewSignalMessage(
		sessionVersion,
		chainKey.Index(),
		previousCounter,
		messageKeys.MacKey(),
		senderEphemeral,
		ciphertextBody,
		sessionState.LocalIdentityKey(),
		sessionState.RemoteIdentityKey(),
	)
	if err != nil {
		return nil, err
	}

	// If we haven't established a session with the recipient yet,
	// send our message as a PreKeySignalMessage.
	if sessionState.HasUnacknowledgedPreKeyMessage() {
		items, err := sessionState.UnackPreKeyMessageItems()
		if err != nil {
			return nil, err
		}
		localRegistrationID := sessionState.LocalRegistrationID()

		ciphertextMessage, err = protocol.NewPreKeySignalMessage(
			sessionVersion,
			localRegistrationID,
			items.PreKeyID(),
			items.SignedPreKeyID(),
			items.BaseKey(),
			sessionState.LocalIdentityKey(),
			ciphertextMessage.(*protocol.SignalMessage),
		)
		if err != nil {
			return nil, err
		}
	}

	sessionState.SetSenderChainKey(chainKey.NextKey())
	if !d.identityKeyStore.IsTrustedIdentity(d.remoteAddress, sessionState.RemoteIdentityKey()) {
		// return err
	}
	d.identityKeyStore.SaveIdentity(d.remoteAddress, sessionState.RemoteIdentityKey())
	d.sessionStore.StoreSession(d.remoteAddress, sessionRecord)
	return ciphertextMessage, nil
}

// Decrypt decrypts the given message using an existing session that
// is stored in the session store.
func (d *Cipher) Decrypt(ciphertextMessage *protocol.SignalMessage) ([]byte, error) {
	plaintext, _, err := d.DecryptAndGetKey(ciphertextMessage)

	return plaintext, err
}

// DecryptAndGetKey decrypts the given message using an existing session that
// is stored in the session store and returns the message keys used for encryption.
func (d *Cipher) DecryptAndGetKey(ciphertextMessage *protocol.SignalMessage) ([]byte, *message.Keys, error) {
	if !d.sessionStore.ContainsSession(d.remoteAddress) {
		return nil, nil, errors.New("No session for: " + d.remoteAddress.String())
	}

	// Load the session record from our session store and decrypt the message.
	sessionRecord, e := d.sessionStore.LoadSession(d.remoteAddress)
	if e != nil {
		return nil, nil, e
	}
	plaintext, messageKeys, err := d.DecryptWithRecord(sessionRecord, ciphertextMessage)
	if err != nil {
		return nil, nil, err
	}

	if !d.identityKeyStore.IsTrustedIdentity(d.remoteAddress, sessionRecord.SessionState().RemoteIdentityKey()) {
		// return err
	}
	d.identityKeyStore.SaveIdentity(d.remoteAddress, sessionRecord.SessionState().RemoteIdentityKey())

	// Store the session record in our session store.
	d.sessionStore.StoreSession(d.remoteAddress, sessionRecord)
	return plaintext, messageKeys, nil
}

func (d *Cipher) DecryptMessage(ciphertextMessage *protocol.PreKeySignalMessage) ([]byte, error) {
	plaintext, _, err := d.DecryptMessageReturnKey(ciphertextMessage)
	return plaintext, err
}

func (d *Cipher) DecryptMessageReturnKey(ciphertextMessage *protocol.PreKeySignalMessage) ([]byte, *message.Keys, error) {
	// Load or create session record for this session.
	sessionRecord, e := d.sessionStore.LoadSession(d.remoteAddress)
	if e != nil {
		return nil, nil, e
	}
	unsignedPreKeyID, err := d.builder.Process(sessionRecord, ciphertextMessage)
	if err != nil {
		return nil, nil, err
	}
	plaintext, keys, err := d.DecryptWithRecord(sessionRecord, ciphertextMessage.WhisperMessage())
	if err != nil {
		return nil, nil, err
	}
	// Store the session record in our session store.
	d.sessionStore.StoreSession(d.remoteAddress, sessionRecord)
	if unsignedPreKeyID != nil && !unsignedPreKeyID.IsEmpty {
		d.preKeyStore.RemovePreKey(unsignedPreKeyID.Value)
	}
	return plaintext, keys, nil
}

// DecryptWithKey will decrypt the given message using the given symmetric key. This
// can be used when decrypting messages at a later time if the message key was saved.
func (d *Cipher) DecryptWithKey(ciphertextMessage *protocol.SignalMessage, key *message.Keys) ([]byte, error) {
	plaintext, err := decrypt(key, ciphertextMessage.Body())
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DecryptWithRecord decrypts the given message using the given session record.
func (d *Cipher) DecryptWithRecord(sessionRecord *record.Session, ciphertext *protocol.SignalMessage) ([]byte, *message.Keys, error) {
	previousStates := sessionRecord.PreviousSessionStates()
	sessionState := sessionRecord.SessionState()

	// Try and decrypt the message with the current session state.
	plaintext, messageKeys, err := d.DecryptWithState(sessionState, ciphertext)

	// If we received an error using the current session state, loop
	// through all previous states.
	if err != nil {
		for i, state := range previousStates {
			// Try decrypting the message with previous states
			plaintext, messageKeys, err = d.DecryptWithState(state, ciphertext)
			if err != nil {
				continue
			}

			// If successful, remove and promote the state.
			previousStates = append(previousStates[:i], previousStates[i+1:]...)
			sessionRecord.PromoteState(state)

			return plaintext, messageKeys, nil
		}

		return nil, nil, errors.New("No valid sessions.")
	}

	// If decryption was successful, set the session state and return the plain text.
	sessionRecord.SetState(sessionState)

	return plaintext, messageKeys, nil
}

// DecryptWithState decrypts the given message with the given session state.
func (d *Cipher) DecryptWithState(sessionState *record.State, ciphertextMessage *protocol.SignalMessage) ([]byte, *message.Keys, error) {
	if !sessionState.HasSenderChain() {
		err := "Uninitialized session!"
		return nil, nil, errors.New(err)
	}

	if ciphertextMessage.MessageVersion() != sessionState.Version() {
		err := "Wrong message version!"
		return nil, nil, errors.New(err)
	}

	messageVersion := ciphertextMessage.MessageVersion()
	theirEphemeral := ciphertextMessage.SenderRatchetKey()
	counter := ciphertextMessage.Counter()
	chainKey, chainCreateErr := getOrCreateChainKey(sessionState, theirEphemeral)
	if chainCreateErr != nil {
		return nil, nil, chainCreateErr
	}

	messageKeys, keysCreateErr := getOrCreateMessageKeys(sessionState, theirEphemeral, chainKey, counter)
	if keysCreateErr != nil {
		return nil, nil, keysCreateErr
	}

	err := ciphertextMessage.VerifyMac(messageVersion, sessionState.RemoteIdentityKey(), sessionState.LocalIdentityKey(), messageKeys.MacKey())
	if err != nil {
		return nil, nil, err
	}

	plaintext, err := d.DecryptWithKey(ciphertextMessage, messageKeys)
	if err != nil {
		return nil, nil, err
	}

	sessionState.ClearUnackPreKeyMessage()

	return plaintext, messageKeys, nil
}

func getOrCreateMessageKeys(sessionState *record.State, theirEphemeral ecc.ECPublicKeyable,
	chainKey *chain.Key, counter uint32) (*message.Keys, error) {

	if chainKey.Index() > counter {
		if sessionState.HasMessageKeys(theirEphemeral, counter) {
			return sessionState.RemoveMessageKeys(theirEphemeral, counter), nil
		}
		index := strconv.FormatUint(uint64(chainKey.Index()), 10)
		count := strconv.FormatUint(uint64(counter), 10)
		return nil, errors.New(
			"Received message with old counter: " + index + " , " + count,
		)
	}

	if counter-chainKey.Index() > maxFutureMessages {
		return nil, errors.New("Too many messages into the future!")
	}

	for chainKey.Index() < counter {
		messageKeys := chainKey.MessageKeys()
		sessionState.SetMessageKeys(theirEphemeral, messageKeys)
		chainKey = chainKey.NextKey()
	}

	sessionState.SetReceiverChainKey(theirEphemeral, chainKey.NextKey())
	return chainKey.MessageKeys(), nil
}

// getOrCreateChainKey will either return the existing chain key or
// create a new one with the given session state and ephemeral key.
func getOrCreateChainKey(sessionState *record.State, theirEphemeral ecc.ECPublicKeyable) (*chain.Key, error) {

	// If our session state already has a receiver chain, use their
	// ephemeral key in the existing chain.
	if sessionState.HasReceiverChain(theirEphemeral) {
		return sessionState.ReceiverChainKey(theirEphemeral), nil
	}

	// If we don't have a chain key, create one with ephemeral keys.
	rootKey := sessionState.RootKey()
	ourEphemeral := sessionState.SenderRatchetKeyPair()
	receiverChain, rErr := rootKey.CreateChain(theirEphemeral, ourEphemeral)
	if rErr != nil {
		return nil, rErr
	}

	// Generate a new ephemeral key pair.
	ourNewEphemeral, gErr := ecc.GenerateKeyPair()
	if gErr != nil {
		return nil, gErr
	}

	// Create a new chain using our new ephemeral key.
	senderChain, cErr := receiverChain.RootKey.CreateChain(theirEphemeral, ourNewEphemeral)
	if cErr != nil {
		return nil, cErr
	}

	// Set our session state parameters.
	sessionState.SetRootKey(senderChain.RootKey)
	sessionState.AddReceiverChain(theirEphemeral, receiverChain.ChainKey)
	previousCounter := max(sessionState.SenderChainKey().Index()-1, 0)
	sessionState.SetPreviousCounter(previousCounter)
	sessionState.SetSenderChain(ourNewEphemeral, senderChain.ChainKey)

	return receiverChain.ChainKey.(*chain.Key), nil
}

// decrypt will use the given message keys and ciphertext and return
// the plaintext bytes.
func decrypt(keys *message.Keys, body []byte) ([]byte, error) {
	return cipher.DecryptCbc(keys.Iv(), keys.CipherKey(), bytehelper.CopySlice(body))
}

// encrypt will use the given cipher, message keys, and plaintext bytes
// and return ciphertext bytes.
func encrypt(messageKeys *message.Keys, plaintext []byte) ([]byte, error) {
	return cipher.EncryptCbc(messageKeys.Iv(), messageKeys.CipherKey(), plaintext)
}

// Max is a uint32 implementation of math.Max
func max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}
