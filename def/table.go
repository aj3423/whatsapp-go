package def

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	Nick   string
	Avatar []byte

	BizAddress     string
	BizDescription string
	BizCategory    string
}

type Device struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	Ja3Config string

	IsBusiness               bool
	HasGooglePlay            bool
	GooglePlayServiceVersion string

	AndroidVersion string // 9
	Brand          string // Xiaomi
	Model          string // MI 6X
	Locale         string // CN
	Language       string // zh

	NetType     int32 // 1:wifi
	NetSubType  int32 // 1:wifi
	MemoryClass int32 // 256

	AndroidApiLevel          int32  // 28
	HasSdCard                int32  // 1
	IsSdCardRemovable        int32  // 0
	CpuAbi                   string // arm64-v8a
	ExternalStorageAvailSize int64
	ExternalStorageTotalSize int64
	StorageAvailSize         int64
	StorageTotalSize         int64
	YearClass                int32
	YearClass2016            int32

	Cc                  string // 86
	Phone               string // 18511112222
	Mcc                 string //
	Mnc                 string //
	SimMcc              string //
	SimMnc              string //
	SimOperatorName     string
	NetworkOperatorName string
	Product             string //
	Build               string //
	Board               string //
	CpuCount            int    //

	ExpId         []byte // scanf from uuid4()
	RegId         uint32
	Fdid          string // uuid4
	BackupToken   []byte // stores encrypted in backup_token, 32bytes rand
	RecoveryToken []byte // stores encrypted in rc2, 32bytes rand
}

type Config struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	RoutingInfo []byte

	VNameCert []byte

	ConnectionLC int // +1 on every login

	IsPassiveActive      bool
	AbPropsConfigKey     string // "1TH,CY,2J,3c"
	AbPropsHash          string // "1rgPhu"
	ServerPropsConfigKey string // "0,CY,2J,3c"
	ServerPropsHash      string // "1xz0eE"

	// Noise
	NoiseLocation string
	StaticPub     []byte
	StaticPriv    []byte
	RemoteStatic  []byte // get from HandshakeXX
}

type Schedule struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	IsGroupsDirty        bool
	IsAccountSyncDirty   bool
	SafetynetAttestation bool // server push ib safetynet
	SafetynetVerifyApps  bool

	GetPropsW                time.Time
	GetPropsAbt              time.Time
	CreateGoogle             time.Time
	ListGroup                time.Time
	GetWbList                time.Time
	GetProfilePicturePreview time.Time
	SetBackupToken           time.Time
	SetEncrypt               time.Time
	GetStatusPrivacy         time.Time
	GetLinkedAccouts         time.Time
	ThriftQueryCatkit        time.Time
	GetBlockList             time.Time
	UsyncDevice              time.Time

	SetBizVerifiedName time.Time
	GetBizVerifiedName time.Time
	GetBizCatalog      time.Time
	//
	SetBizProfile     time.Time
	GetBizProfile_4   time.Time
	GetBizProfile_116 time.Time

	GetStatusUser       time.Time
	GetDisappearingMode time.Time
	GetPrivacy          time.Time
	GetProfilePicture   time.Time
	GetJabberIqPrivacy  time.Time
	AvailableNick       time.Time
	BizBlockReason      time.Time

	Wam time.Time // last time sent w:stats
}

type Session struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId       uint64
	RecipientId uint
	DeviceId    uint32

	Record []byte
}

type Prekey struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId    uint64
	PrekeyId uint32

	Record    []byte
	CreatedAt time.Time
	DeletedAt time.Time
}

type Identity struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId       uint64
	RecipientId uint
	DeviceId    uint32

	PublicKey    []byte
	PrivateKey   []byte
	NextPrekeyId uint32
}

type SignedPrekey struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId    uint64
	PrekeyId uint32

	Record []byte
}
type SenderKey struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId    uint64
	GroupId  string
	SenderId string
	DeviceId uint32

	Record []byte
}

type Proxy struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	Addr string
	Dns  map[string]string
}

type Message struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64
	MsgId string

	Node []byte

	MediaType uint32 // text, image, sticker, ...

	Decrypted bool
	DecMedia  []byte

	RetryTimes    uint32
	RetryPrekeyId uint32

	UpdatedAt time.Time
}

type Group struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64
	Gid   string

	Creator string
	Subject string
}
type GroupMember struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId   uint64
	GroupId string
	Jid     string
}
type WamSchedule struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	//WamPsIdCreate time.Time // 2310
	WamDaily    time.Time // 1158
	WamPttDaily time.Time // 2938
	// reg daily schedule
	WamRegistrationComplete                time.Time // 1342
	WamAndroidDatabaseMigrationDailyStatus time.Time // 2318
	WamStatusDaily                         time.Time // 1676
	WamAndroidDatabaseMigrationEvent       time.Time // 1912
	WamSmbVnameCertHealth                  time.Time // 1602
}
type WamEvent struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64
	ReqId uint16

	LastEventTime time.Time

	EulaAccept           time.Time // inited at first AccOn
	RegFillBizInfoScreen int32     // milli seconds

	// used in WamDaily
	AddressBookSize   int32
	AddressBookWASize int32
	ChatDatabaseSize  int32

	Buffer [][]byte
}

// CDN, 1 Record
type Cdn struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId uint64

	UploadTokenRandomBytes []byte

	MediaConnId uint
	Auth        string
	MaxBuckets  int

	Video    string
	Image    string
	Gif      string
	Ptt      string
	Sticker  string
	Document string
}

type MultiDevice struct {
	ID primitive.ObjectID `bson:"_id"`

	AccId    uint64
	RecId    uint64
	DeviceId uint32

	LastSync time.Time
}
