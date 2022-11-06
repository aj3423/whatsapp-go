package clone

// for clone Phone acc -> bot

type Gene struct {
	Nick                  string
	Fdid                  string
	RoutingInfo           []byte
	NoiseLocation         string
	AbPropsConfigKey      string
	AbPropsConfigHash     string
	ServerPropsConfigKey  string
	ServerPropsConfigHash string

	Prekeys []Prekey
	Identity
	SignedPrekey

	StaticPub  []byte
	StaticPriv []byte
	//RemoteStatic []byte
}

type Prekey struct {
	PrekeyId           int
	SentToServer       bool
	Record             []byte
	DirectDistribution bool
}
type Identity struct {
	RecipientId    int
	DeviceId       int
	RegistrationId int
	PublicKey      []byte
	PrivateKey     []byte
	NextPrekeyId   int
}
type SignedPrekey struct {
	PrekeyId int
	Record   []byte
}
