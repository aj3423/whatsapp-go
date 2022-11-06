package protocol

type CiphertextMessage interface {
	Serialize() []byte
	Type() uint32
}

const (
	UnsupportedVersion = 1
	CurrentVersion     = 3
)

const (
	WHISPER_TYPE                = 2
	PREKEY_TYPE                 = 3
	SENDERKEY_TYPE              = 4
	SENDERKEY_DISTRIBUTION_TYPE = 5
)
