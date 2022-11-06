package db

import (
	"bytes"
	"strconv"
	"strings"
	"sync"
	"time"

	groupRecord "wa/signal/groups/state/record"

	"ahex"
	"ajson"
	"algo/xed25519"
	"arand"
	"wa/crypto"
	"wa/def"
	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/protocol"
	"wa/signal/state/record"
	"wa/signal/util/bytehelper"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB_NAME = `wa`

type Store struct {
	acc_id uint64

	muSession  sync.Mutex
	muWamEvent sync.RWMutex
}

/*
add a col:
 1. add colXXX
 2. add index
 3. add in DeleteAcc
*/
var colProfile *mongo.Collection
var colDevice *mongo.Collection
var colConfig *mongo.Collection
var colSchedule *mongo.Collection
var colProxy *mongo.Collection
var colSession *mongo.Collection
var colPrekey *mongo.Collection
var colIdentity *mongo.Collection
var colSignedPrekey *mongo.Collection
var colSenderKey *mongo.Collection
var colMessage *mongo.Collection
var colGroup *mongo.Collection
var colGroupMember *mongo.Collection
var colWamSchedule *mongo.Collection
var colWamEvent *mongo.Collection
var colCdn *mongo.Collection
var colMultiDevice *mongo.Collection

func init() {
	colProfile = client.Database(DB_NAME).Collection(`Profile`)
	colDevice = client.Database(DB_NAME).Collection(`Device`)
	colConfig = client.Database(DB_NAME).Collection(`Config`)
	colSchedule = client.Database(DB_NAME).Collection(`Schedule`)
	colProxy = client.Database(DB_NAME).Collection(`Proxy`)
	colSession = client.Database(DB_NAME).Collection(`Session`)
	colPrekey = client.Database(DB_NAME).Collection(`Prekey`)
	colIdentity = client.Database(DB_NAME).Collection(`Identity`)
	colSignedPrekey = client.Database(DB_NAME).Collection(`SignedPrekey`)
	colSenderKey = client.Database(DB_NAME).Collection(`SenderKey`)
	colMessage = client.Database(DB_NAME).Collection(`Message`)
	colGroup = client.Database(DB_NAME).Collection(`Group`)
	colGroupMember = client.Database(DB_NAME).Collection(`GroupMember`)
	colWamSchedule = client.Database(DB_NAME).Collection(`WamSchedule`)
	colWamEvent = client.Database(DB_NAME).Collection(`WamEvent`)
	colCdn = client.Database(DB_NAME).Collection(`Cdn`)
	colMultiDevice = client.Database(DB_NAME).Collection(`MultiDevice`)

	_, e1 := colProfile.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e2 := colDevice.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e3 := colConfig.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e4 := colSchedule.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e5 := colSession.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: `RecipientId`, Value: 1},
			{Key: `DeviceId`, Value: 1},
		},
	})
	_, e6 := colPrekey.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "PrekeyId", Value: 1},
		},
	})
	_, e7 := colIdentity.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "RecipientId", Value: 1},
			{Key: "DeviceId", Value: 1},
		},
	})
	_, e8 := colSignedPrekey.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "PrekeyId", Value: 1},
		},
	})
	_, e9 := colSenderKey.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1},
	})
	_, e10 := colProxy.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e11 := colMessage.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "MsgId", Value: 1},
		},
	})
	_, e12 := colGroup.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "Gid", Value: 1},
		},
	})
	_, e13 := colGroupMember.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "Groupid", Value: 1},
			{Key: "Jid", Value: 1},
		},
	})
	_, e14 := colWamSchedule.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e15 := colWamEvent.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})
	_, e16 := colCdn.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{"AccId": 1}, Options: options.Index().SetUnique(true)})

	_, e17 := colMultiDevice.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "AccId", Value: 1},
			{Key: "RecId", Value: 1},
			{Key: "DeviceId", Value: 1},
		},
	})

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || e6 != nil || e7 != nil || e8 != nil || e9 != nil || e10 != nil || e11 != nil || e12 != nil || e13 != nil || e14 != nil || e15 != nil || e16 != nil || e17 != nil {
		panic(`fail create db index`)
	}
}

func NewStore(acc_id uint64) *Store {
	return &Store{
		acc_id: acc_id,
	}
}
func DeleteAcc(acc_id uint64) error {
	_, e1 := colProfile.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e2 := colDevice.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e3 := colConfig.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e4 := colSchedule.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e5 := colProxy.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e6 := colSession.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e7 := colPrekey.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e8 := colIdentity.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e9 := colSignedPrekey.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e10 := colSenderKey.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e11 := colMessage.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e12 := colGroup.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e13 := colGroupMember.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e14 := colWamSchedule.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e15 := colCdn.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e16 := colWamEvent.DeleteMany(ctx, bson.M{`AccId`: acc_id})
	_, e17 := colMultiDevice.DeleteMany(ctx, bson.M{`AccId`: acc_id})

	_, e18 := colLog.DeleteMany(ctx, bson.M{`AccId`: acc_id})

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || e6 != nil || e7 != nil || e8 != nil || e9 != nil || e10 != nil || e11 != nil || e12 != nil || e13 != nil || e14 != nil || e15 != nil || e16 != nil || e17 != nil || e18 != nil {
		return errors.New(`db Delete err`)
	}

	return nil
}

func (s *Store) GetProfile() (*def.Profile, error) {
	prof := &def.Profile{}
	e := colProfile.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(prof)
	if errors.Is(e, mongo.ErrNoDocuments) {
		_, e = colProfile.InsertOne(ctx, bson.M{`AccId`: s.acc_id})
	}
	if e != nil {
		return nil, errors.Wrap(e, `fail get profile`)
	}
	return prof, nil
}
func (s *Store) ModifyProfile(mod bson.M) error {
	cur := colProfile.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return cur.Err()
}

func (s *Store) GetDev() (*def.Device, error) {
	dev := &def.Device{}
	e := colDevice.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(dev)
	if e != nil {
		return nil, errors.Wrap(e, `fail get dev`)
	}
	return dev, nil
}
func (s *Store) ModifyDev(mod bson.M) error {
	cur := colDevice.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return cur.Err()
}

func (s *Store) GetMyJid() (string, error) {
	dev, e := s.GetDev()
	if e != nil {
		return ``, e
	}
	return dev.Cc + dev.Phone + `@s.whatsapp.net`, nil
}

func (s *Store) SetJsonDev(j *ajson.Json) error {
	must := true // default true, must have some fields
	if v, e := j.Get(`validate`).TryBool(); e == nil {
		must = v
	}

	mod := bson.M{}
	// --------- attributes ------------
	if AndroidVersion, e := j.Get(`AndroidVersion`).TryString(); e == nil {
		mod[`AndroidVersion`] = AndroidVersion
	} else if must {
		return errors.New(`missing "AndroidVersion"`)
	}
	if Brand, e := j.Get(`Brand`).TryString(); e == nil {
		mod[`Brand`] = Brand
	} else if must {
		return errors.New(`missing "Brand"`)
	}
	if Model, e := j.Get(`Model`).TryString(); e == nil {
		mod[`Model`] = Model
	} else if must {
		return errors.New(`missing "Model"`)
	}
	if Locale, e := j.Get(`Locale`).TryString(); e == nil {
		mod[`Locale`] = Locale
	} else if must {
		return errors.New(`missing "Locale"`)
	}
	if Language, e := j.Get(`Language`).TryString(); e == nil {
		mod[`Language`] = Language
	} else if must {
		return errors.New(`missing "Language"`)
	}
	if NetType, e := j.Get(`NetType`).TryInt(); e == nil {
		mod[`NetType`] = NetType
	} else if must {
		return errors.New(`missing "NetType"`)
	}
	if NetSubType, e := j.Get(`NetSubType`).TryInt(); e == nil {
		mod[`NetSubType`] = NetSubType
	} else if must {
		return errors.New(`missing "NetSubType"`)
	}
	if MemoryClass, e := j.Get(`MemoryClass`).TryInt(); e == nil {
		mod[`MemoryClass`] = MemoryClass
	} else if must {
		return errors.New(`missing "MemoryClass"`)
	}
	if AndroidApiLevel, e := j.Get(`AndroidApiLevel`).TryInt(); e == nil {
		mod[`AndroidApiLevel`] = AndroidApiLevel
	} else if must {
		return errors.New(`missing "AndroidApiLevel"`)
	}
	if HasSdCard, e := j.Get(`HasSdCard`).TryInt(); e == nil {
		mod[`HasSdCard`] = HasSdCard
	} else if must {
		return errors.New(`missing "HasSdCard"`)
	}
	if IsSdCardRemovable, e := j.Get(`IsSdCardRemovable`).TryInt(); e == nil {
		mod[`IsSdCardRemovable`] = IsSdCardRemovable
	} else if must {
		return errors.New(`missing "IsSdCardRemovable"`)
	}
	if CpuAbi, e := j.Get(`CpuAbi`).TryString(); e == nil {
		mod[`CpuAbi`] = CpuAbi
	} else if must {
		return errors.New(`missing "CpuAbi"`)
	}
	if YearClass, e := j.Get(`YearClass`).TryInt(); e == nil {
		mod[`YearClass`] = YearClass
	}
	if YearClass2016, e := j.Get(`YearClass2016`).TryInt(); e == nil {
		mod[`YearClass2016`] = YearClass2016
	}
	if ExternalStorageAvailSize, e := j.Get(`ExternalStorageAvailSize`).TryInt64(); e == nil {
		mod[`ExternalStorageAvailSize`] = ExternalStorageAvailSize
	} else if must {
		return errors.New(`missing "ExternalStorageAvailSize"`)
	}
	if ExternalStorageTotalSize, e := j.Get(`ExternalStorageTotalSize`).TryInt64(); e == nil {
		mod[`ExternalStorageTotalSize`] = ExternalStorageTotalSize
	} else if must {
		return errors.New(`missing "ExternalStorageTotalSize"`)
	}
	if StorageAvailSize, e := j.Get(`StorageAvailSize`).TryInt64(); e == nil {
		mod[`StorageAvailSize`] = StorageAvailSize
	} else if must {
		return errors.New(`missing "StorageAvailSize"`)
	}
	if StorageTotalSize, e := j.Get(`StorageTotalSize`).TryInt64(); e == nil {
		mod[`StorageTotalSize`] = StorageTotalSize
	} else if must {
		return errors.New(`missing "StorageTotalSize"`)
	}
	if Cc, e := j.Get(`Cc`).TryString(); e == nil {
		mod[`Cc`] = Cc
	} else if must {
		return errors.New(`missing "Cc"`)
	}
	if Phone, e := j.Get(`Phone`).TryString(); e == nil {
		mod[`Phone`] = Phone
	} else if must {
		return errors.New(`missing "Phone"`)
	}
	if Mcc, e := j.Get(`Mcc`).TryString(); e == nil {
		mod[`Mcc`] = Mcc
	} else if must {
		return errors.New(`missing "Mcc"`)
	}
	if Mnc, e := j.Get(`Mnc`).TryString(); e == nil {
		mod[`Mnc`] = Mnc
	} else if must {
		return errors.New(`missing "Mnc"`)
	}
	if SimMcc, e := j.Get(`SimMcc`).TryString(); e == nil {
		mod[`SimMcc`] = SimMcc
	} else if must {
		return errors.New(`missing "SimMcc"`)
	}
	if SimMnc, e := j.Get(`SimMnc`).TryString(); e == nil {
		mod[`SimMnc`] = SimMnc
	} else if must {
		return errors.New(`missing "SimMnc"`)
	}
	if SimOperatorName, e := j.Get(`SimOperatorName`).TryString(); e == nil {
		mod[`SimOperatorName`] = SimOperatorName
	} else if must {
		return errors.New(`missing "SimOperatorName"`)
	}
	if NetworkOperatorName, e := j.Get(`NetworkOperatorName`).TryString(); e == nil {
		mod[`NetworkOperatorName`] = NetworkOperatorName
	} else if must {
		return errors.New(`missing "NetworkOperatorName"`)
	}
	if Product, e := j.Get(`Product`).TryString(); e == nil {
		mod[`Product`] = Product
	} else if must {
		return errors.New(`missing "Product"`)
	}
	if Build, e := j.Get(`Build`).TryString(); e == nil {
		mod[`Build`] = Build
	} else if must {
		return errors.New(`missing "Build"`)
	}
	if Board, e := j.Get(`Board`).TryString(); e == nil {
		mod[`Board`] = Board
	} else if must {
		return errors.New(`missing "Board"`)
	}
	if CpuCount, e := j.Get(`CpuCount`).TryString(); e == nil {
		mod[`CpuCount`] = CpuCount
	} else {
		mod[`CpuCount`] = 8 // default 8
	}
	if IsBusiness, e := j.Get(`IsBusiness`).TryBool(); e == nil {
		mod[`IsBusiness`] = IsBusiness
	}
	if HasGooglePlay, e := j.Get(`HasGooglePlay`).TryBool(); e == nil {
		mod[`HasGooglePlay`] = HasGooglePlay
	}
	if GooglePlayServiceVersion, e := j.Get(`GooglePlayServiceVersion`).TryString(); e == nil {
		mod[`GooglePlayServiceVersion`] = GooglePlayServiceVersion
	}

	if ja3_type, e := j.Get(`Ja3Config`).Get(`type`).TryInt(); e == nil {
		str := ``

		switch int8(ja3_type) {
		case def.Ja3Type_Default:
			str = def.Ja3_WhiteMi6x
		case def.Ja3Type_RandomGen:
			str = def.NewRandomJa3Config()
		case def.Ja3Type_Custom:
			str, e = j.Get(`Ja3Config`).Get(`value`).TryString()
			if e != nil {
				return errors.New(`missing Ja3Config.value`)
			}
		}

		mod[`Ja3Config`] = str
	}

	return s.ModifyDev(mod)
}

func (s *Store) GetProxy() (string, map[string]string, error) {
	prx := &def.Proxy{}
	e := colProxy.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(prx)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return ``, nil, nil
	}

	return prx.Addr, prx.Dns, errors.Wrap(e, `fail get proxy`)
}
func (s *Store) GetDns() (map[string]string, error) {
	prx := &def.Proxy{}
	e := colProxy.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(prx)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return nil, nil
	}

	return prx.Dns, errors.Wrap(e, `fail get dns`)
}
func (s *Store) SetProxy(addr string) error {
	r := colProxy.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: bson.M{`Addr`: addr},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
func (s *Store) SetDns(dns map[string]string) error {
	r := colProxy.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: bson.M{`Dns`: dns},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) AccExists() (bool, error) {
	iden := &def.Identity{}
	e := colIdentity.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(iden)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return false, nil
	}
	if e != nil {
		return false, errors.Wrap(e, `fail check acc exists`)
	}
	return true, nil
}

// if Creating New Account, initialize keys:
// 1. Device: ExpId, RegId, Fdid, BackupToken, RecoveryToken
// 2. Identity
// 3. SignedPrekey
// 4. NoiseStatic
func (s *Store) FirstInit() error {

	var iden_priv, iden_pub []byte
	{ // Identity
		iden_priv, iden_pub = crypto.NewECKeyPair()
		iden := identity.NewKeyPair(
			identity.NewKeyFromBytes(bytehelper.SliceToArray(iden_pub)),
			ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(iden_priv)),
		)

		e := s.SetMyIdentityKeyPair(iden)
		if e != nil {
			return e
		}
		e = s.SetMyNextPrekeyId(uint32(arand.Int(0x100000, 0x600000)))
		if e != nil {
			return e
		}
	}

	{ // SignedPrekey
		kp, e := ecc.GenerateKeyPair()
		if e != nil {
			return e
		}
		id := uint32(0)

		signature, e := xed25519.Sign(iden_priv, kp.PublicKey().Serialize())
		if e != nil {
			return e
		}
		record := record.NewSignedPreKey(
			id,
			time.Now().UnixMilli(),
			kp,
			bytehelper.SliceToArray64(signature))

		e = s.StoreSignedPreKey(id, record)
		if e != nil {
			return e
		}
	}
	{ //Config
		if e := s.GenerateNoiseStatic(); e != nil {
			return e
		}
	}

	// ExpId, RegId, Fdid, BackupToken, RecoveryToken
	{
		e := s.ModifyDev(bson.M{
			`Ja3Config`:     def.Ja3_WhiteMi6x,
			`ExpId`:         ahex.Dec(strings.ReplaceAll(arand.Uuid4(), `-`, ``)),
			`Fdid`:          arand.Uuid4(),
			`RegId`:         uint32(arand.Int(0x1, 0x7ffffffd)),
			`BackupToken`:   arand.Bytes(20), // 20 bytes
			`RecoveryToken`: arand.Bytes(20), // 20 bytes
		})
		if e != nil {
			return e
		}
	}

	// cdn
	{
		e := s.ModifyCdn(bson.M{
			`UploadTokenRandomBytes`: arand.Bytes(0x20),
		})
		if e != nil {
			return e
		}
	}
	return nil
}

/*
--------- Schedule ----------
*/
func (s *Store) GetSchedule() (*def.Schedule, error) {
	sch := &def.Schedule{}

	e := colSchedule.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(sch)

	if errors.Is(e, mongo.ErrNoDocuments) {
		_, e = colSchedule.InsertOne(ctx, bson.M{`AccId`: s.acc_id})
	}
	return sch, errors.Wrap(e, `fail get schedule`)
}
func (s *Store) ModifySchedule(mod bson.M) error {
	r := colSchedule.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
func (s *Store) SetGroupsDirty() error {
	return s.ModifySchedule(bson.M{
		`IsGroupDirty`: true,
	})
}
func (s *Store) SetAccountSyncDirty() error {
	return s.ModifySchedule(bson.M{
		`IsAccountSyncDirty`: true,
	})
}
func (s *Store) SetSafetynetAttestation() error {
	return s.ModifySchedule(bson.M{
		`SafetynetAttestation`: true,
	})
}
func (s *Store) SetSafetynetVerifyApps() error {
	return s.ModifySchedule(bson.M{
		`SafetynetVerifyApps`: true,
	})
}

// Media
func (s *Store) GetCdn() (*def.Cdn, error) {
	cdn := &def.Cdn{}

	e := colCdn.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(cdn)

	if errors.Is(e, mongo.ErrNoDocuments) {
		_, e = colCdn.InsertOne(ctx, bson.M{`AccId`: s.acc_id})
	} else if e != nil {
		return nil, errors.Wrap(e, `fail get cdn`)
	}
	return cdn, e
}
func (s *Store) GetMediaConnId() (uint, error) {
	cdn, e := s.GetCdn()
	if e != nil {
		return 0, e
	}
	return cdn.MediaConnId, nil
}
func (s *Store) ModifyCdn(mod bson.M) error {
	cur := colCdn.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return cur.Err()
}

// multi device
func (s *Store) GetMultiDevice(recid uint64) ([]uint32, error) {
	var devices []uint32

	cur, e := colMultiDevice.Find(ctx, bson.M{
		`AccId`: s.acc_id,
		`RecId`: recid,
	})
	if e != nil {
		return nil, e
	}
	for cur.Next(ctx) {
		x := &def.MultiDevice{}
		e := cur.Decode(x)
		if e != nil {
			return nil, e
		}
		devices = append(devices, x.DeviceId)
	}
	return devices, nil
}
func (s *Store) AddMultiDevice(recid uint64, devId uint32) error {
	mod := bson.M{
		`AccId`:    s.acc_id,
		`RecId`:    recid,
		`DeviceId`: devId,
	}

	r := colMultiDevice.FindOneAndUpdate(ctx, mod, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	return r.Err()
}
func (s *Store) DelMultiDevice(recid uint64, devId uint32) error {
	_, e := colMultiDevice.DeleteOne(ctx, bson.M{
		`AccId`:    s.acc_id,
		`RecId`:    recid,
		`DeviceId`: devId,
	})
	return e
}
func (s *Store) DelAllMultiDevice(recid uint64) error {
	_, e := colMultiDevice.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id,
		`RecId`: recid,
	})
	return e
}
func (s *Store) GetMultiDeviceLastSync(recid uint64, devid uint32) (time.Time, error) {
	mds := &def.MultiDevice{}

	e := colMultiDevice.FindOne(ctx, bson.M{
		`AccId`:    s.acc_id,
		`RecId`:    recid,
		`DeviceId`: devid,
	}).Decode(mds)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return time.Time{}, nil
	}
	if e != nil {
		return time.Time{}, errors.Wrap(e, `fail get mds`)
	}

	return mds.LastSync, nil
}
func (s *Store) SetMultiDeviceLastSync(recid uint64, devid uint32) error {
	r := colMultiDevice.FindOneAndUpdate(ctx, bson.M{
		`AccId`:    s.acc_id,
		`RecId`:    recid,
		`DeviceId`: devid,
	}, bson.M{
		`$set`: bson.M{
			`LastSync`: time.Now(),
		},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) SaveNoiseLocation(loc string) error {
	return s.ModifyConfig(bson.M{
		`NoiseLocation`: loc,
	})
}
func (s *Store) GetConfig() (*def.Config, error) {
	cfg := &def.Config{}

	e := colConfig.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(cfg)

	if errors.Is(e, mongo.ErrNoDocuments) {
		_, e = colConfig.InsertOne(ctx, bson.M{`AccId`: s.acc_id})
	}
	if e != nil {
		return nil, errors.Wrap(e, `fail get cfg`)
	}

	return cfg, e
}
func (s *Store) ModifyConfig(mod bson.M) error {
	r := colConfig.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

// Noise Static key
func (s *Store) GenerateNoiseStatic() error {
	StaticPriv, StaticPub := crypto.NewECKeyPair()
	return s.ModifyConfig(bson.M{
		`StaticPriv`: StaticPriv,
		`StaticPub`:  StaticPub,
	})
}

func (s *Store) SaveRemoteNoiseStatic(r_pub []byte) error {
	return s.ModifyConfig(bson.M{
		`RemoteStatic`: r_pub,
	})
}

func (s *Store) GetLocalRegistrationId() (uint32, error) {
	d, e := s.GetDev()
	return d.RegId, e
}
func (s *Store) SetLocalRegistrationId(id uint32) error {
	return s.ModifyDev(bson.M{
		`RegId`: id,
	})
}

func (s *Store) ModifyIdentity(filter, mod bson.M) error {
	r := colIdentity.FindOneAndUpdate(ctx, filter, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
func (s *Store) ModifyMyIdentity(mod bson.M) error {
	return s.ModifyIdentity(bson.M{
		`AccId`: s.acc_id, `RecipientId`: 0, `DeviceId`: 0,
	}, mod)
}
func (s *Store) GetMyIdentity() (*def.Identity, error) {
	iden := &def.Identity{}
	e := colIdentity.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: 0, `DeviceId`: 0,
	}).Decode(iden)
	return iden, e
}

// get my idkey pair
func (s *Store) GetIdentityKeyPair() (*identity.KeyPair, error) {
	iden, e := s.GetMyIdentity()
	if e != nil {
		return nil, e
	}
	publicKey := identity.NewKeyFromBytes(bytehelper.SliceToArray(iden.PublicKey))
	privKey := ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(iden.PrivateKey))
	return identity.NewKeyPair(publicKey, privKey), nil
}

// set my idkey pair
func (s *Store) SetMyIdentityKeyPair(ikp *identity.KeyPair) error {
	priv := ikp.PrivateKey().Serialize()
	pub := ikp.PublicKey().PublicKey().PublicKey() // s h i t
	return s.ModifyMyIdentity(bson.M{
		`PrivateKey`: priv[:], `PublicKey`: pub[:],
	})
}

// called by Radical
func (s *Store) SaveIdentity(addr *protocol.SignalAddress, identityKey *identity.Key) error {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return e
	}

	filter := bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	}
	pub := identityKey.PublicKey().PublicKey()
	mod := bson.M{
		`PublicKey`: pub[:],
	}
	return s.ModifyIdentity(filter, mod)
}
func (s *Store) DeleteIdentity(addr *protocol.SignalAddress) error {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return e
	}
	_, e = colIdentity.DeleteOne(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	})
	return e
}
func (s *Store) IsTrustedIdentity(addr *protocol.SignalAddress, identityKey *identity.Key) bool {

	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return false
	}

	iden := &def.Identity{}
	e = colIdentity.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	}).Decode(iden)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return true
	}
	if e != nil {
		return false
	}

	pub := identityKey.PublicKey().PublicKey()
	return bytes.Equal(iden.PublicKey, pub[:])
}

func (s *Store) SetMyNextPrekeyId(prekey_id uint32) error {
	return s.ModifyMyIdentity(bson.M{
		`NextPrekeyId`: prekey_id,
	})
}
func (s *Store) GetMyNextPrekeyId() (uint32, error) {
	iden, e := s.GetMyIdentity()
	if e != nil {
		return 0, e
	}
	next := iden.NextPrekeyId + 1
	return next, s.SetMyNextPrekeyId(next)
}
func (s *Store) GeneratePrekey(prekey_id uint32) (*record.PreKey, error) {
	kp, _ := ecc.GenerateKeyPair()
	rec := record.NewPreKey(prekey_id, kp)

	err := s.StorePreKey(prekey_id, rec)
	return rec, err
}

func (s *Store) LoadPreKey(prekey_id uint32) (*record.PreKey, error) {
	k := &def.Prekey{}
	e := colPrekey.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: prekey_id,
	}).Decode(k)
	if e != nil {
		return nil, e
	}
	return record.NewPreKeyFromBytes(k.Record)
}
func (s *Store) StorePreKey(prekey_id uint32, rec *record.PreKey) error {
	ser, e := rec.Serialize()
	if e != nil {
		return e
	}

	r := colPrekey.FindOneAndUpdate(ctx, bson.M{
		`AccId`:    s.acc_id,
		`PrekeyId`: prekey_id,
	}, bson.M{
		`$set`: bson.M{`Record`: ser},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
func (s *Store) ContainsPreKey(prekey_id uint32) bool {
	k := &def.Prekey{}
	e := colPrekey.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: prekey_id,
	}).Decode(k)
	return e == nil
}
func (s *Store) RemovePreKey(prekey_id uint32) {
	colPrekey.FindOneAndUpdate(ctx, bson.M{
		`AccId`:    s.acc_id,
		`PrekeyId`: prekey_id,
	}, bson.M{
		`$set`: bson.M{`DeletedAt`: time.Now()},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
}

func (s *Store) ModifySignedPrekey(mod bson.M) error {
	r := colSignedPrekey.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) ContainsSignedPreKey(id uint32) bool {
	if id == 0 {
		return true
	}
	spk := &def.SignedPrekey{}
	e := colSignedPrekey.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: id,
	}).Decode(spk)
	return e == nil
}
func (s *Store) LoadSignedPreKey(prekey_id uint32) (*record.SignedPreKey, error) {
	spk := &def.SignedPrekey{}
	e := colSignedPrekey.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: prekey_id,
	}).Decode(spk)
	if e != nil {
		return nil, e
	}
	return record.NewSignedPreKeyFromBytes(spk.Record)
}
func (s *Store) LoadSignedPreKeys() []*record.SignedPreKey {
	keys := []*record.SignedPreKey{}
	// not used by radical
	return keys
}

func (s *Store) StoreSignedPreKey(prekey_id uint32, rec *record.SignedPreKey) error {

	cur := colSignedPrekey.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: prekey_id,
	}, bson.M{
		`$set`: bson.M{
			`Record`: rec.Serialize(),
		},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return cur.Err()
}
func (s *Store) RemoveSignedPreKey(prekey_id uint32) {
	colSignedPrekey.DeleteOne(ctx, bson.M{
		`AccId`: s.acc_id, `PrekeyId`: prekey_id,
	})
}

func (s *Store) Lock() {
	s.muSession.Lock()
}
func (s *Store) Unlock() {
	s.muSession.Unlock()
}
func (s *Store) LoadSession(addr *protocol.SignalAddress) (*record.Session, error) {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return nil, e
	}
	sess := &def.Session{}
	e = colSession.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	}).Decode(sess)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return record.NewSession(), nil
	}
	if e != nil {
		return nil, e
	}

	return record.NewSessionFromBytes(sess.Record)
}
func (s *Store) StoreSession(addr *protocol.SignalAddress, rec *record.Session) error {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return e
	}
	_rec, e := rec.Serialize()
	if e != nil {
		return e
	}
	r := colSession.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	}, bson.M{
		`$set`: bson.M{`Record`: _rec},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
func (s *Store) ContainsSession(addr *protocol.SignalAddress) bool {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return false
	}
	sess := &def.Session{}
	e = colSession.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	}).Decode(sess)

	return e == nil
}
func (s *Store) DeleteSession(addr *protocol.SignalAddress) {
	recid, e := strconv.Atoi(addr.Name())
	if e != nil {
		return
	}
	colSession.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id, `RecipientId`: uint(recid), `DeviceId`: addr.DeviceID(),
	})
}

func (s *Store) DeleteAllSessions() {
	// not used by radical
}

func (s *Store) GetSubDeviceSessions(recipientID string) []uint32 {
	// not used by radical
	return nil
}

// sender key
func (s *Store) StoreSenderKey(
	skn *protocol.SenderKeyName,
	rec *groupRecord.SenderKey,
) error {
	ser, e := rec.Serialize()
	if e != nil {
		return e
	}
	r := colSenderKey.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: skn.GroupID(), `SenderId`: skn.Sender().Name(), `DeviceId`: skn.Sender().DeviceID(),
	}, bson.M{
		`$set`: bson.M{`Record`: ser},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) LoadSenderKey(
	skn *protocol.SenderKeyName,
) (*groupRecord.SenderKey, error) {
	k := &def.SenderKey{}
	e := colSenderKey.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: skn.GroupID(), `SenderId`: skn.Sender().Name(), `DeviceId`: skn.Sender().DeviceID(),
	}).Decode(k)

	if errors.Is(e, mongo.ErrNoDocuments) {
		return groupRecord.NewSenderKey(), nil
	}

	if e != nil {
		return nil, e
	}
	return groupRecord.NewSenderKeyFromBytes(k.Record)
}
func (s *Store) DeleteSenderKey(addr *protocol.SignalAddress) error {
	_, e := colSenderKey.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id, `SenderId`: addr.Name(), `DeviceId`: addr.DeviceID(),
	})
	return e
}

// message
func (s *Store) EnsureMessage(
	msg_id string,
	n []byte,
) (*def.Message, error) {

	r := colMessage.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `MsgId`: msg_id,
	}, bson.M{
		`$set`: bson.M{
			`Node`: n,
		},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	if r.Err() != nil {
		return nil, r.Err()
	}
	m := &def.Message{}
	e := r.Decode(m)
	if e != nil {
		return nil, e
	}

	s.WamSetHasNewMessage(true)
	return m, nil
}
func (s *Store) GetMessage(msg_id string) (*def.Message, error) {
	m := &def.Message{}
	e := colMessage.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `MsgId`: msg_id,
	}).Decode(m)

	return m, e
}
func (s *Store) ModifyMessage(msg_id string, mod bson.M) error {
	cur := colMessage.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `MsgId`: msg_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return cur.Err()
}

func (s *Store) GetMessageRetry(
	msg_id string,
) (uint32, error) {
	m, e := s.GetMessage(msg_id)
	if e != nil {
		return 0, e
	}
	return m.RetryTimes, nil
}
func (s *Store) IncreaseMessageRetry(
	msg_id string,
) error {
	times, e := s.GetMessageRetry(msg_id)
	if e != nil {
		return e
	}
	return s.ModifyMessage(msg_id, bson.M{
		`RetryTimes`: times + 1,
	})
}
func (s *Store) SaveDecodedMessage(
	msg_id string,

	media_type uint32, // text/media/...
	media []byte,
) error {
	return s.ModifyMessage(msg_id, bson.M{
		`Decrypted`: true,
		`MediaType`: media_type,
		`DecMedia`:  media,
	})
}

func (s *Store) ListMessages() ([]*def.Message, error) {
	var ms []*def.Message

	cur, e := colMessage.Find(ctx, bson.M{
		`AccId`: s.acc_id,
	})
	if e != nil {
		return nil, e
	}
	for cur.Next(ctx) {
		x := &def.Message{}
		e := cur.Decode(x)
		if e != nil {
			return nil, e
		}
		ms = append(ms, x)
	}
	return ms, e
}
func (s *Store) DeleteMessage(
	msg_id string,
) error {
	_, e := colMessage.DeleteOne(ctx, bson.M{
		`AccId`: s.acc_id, `MsgId`: msg_id,
	})
	return e
}
func (s *Store) DeleteMessages(
	msg_ids []string,
) error {
	for _, id := range msg_ids {
		e := s.DeleteMessage(id)
		if e != nil {
			return e
		}
	}
	return nil
}
func (s *Store) CreateGroup(
	gid, subject, creator string, members []string,
) error {
	// 1. store group
	r := colGroup.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id, `Gid`: gid,
	}, bson.M{
		`$set`: bson.M{
			`Creator`: creator, `Subject`: subject,
		},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	if r.Err() != nil {
		return r.Err()
	}
	g := &def.Group{}
	e := r.Decode(g)
	if e != nil {
		return e
	}
	// 2. store members
	for _, jid := range members {
		_, e := colGroupMember.InsertOne(ctx, bson.M{
			`AccId`: s.acc_id, `GroupId`: g.ID, `Jid`: jid,
		})
		if e != nil {
			return e
		}
	}
	return nil
}

func (s *Store) GroupCount() (int, error) {
	cnt, e := colGroup.CountDocuments(ctx, bson.M{
		`AccId`: s.acc_id,
	})
	return int(cnt), e
}
func (s *Store) RemoveAllGroups() error {
	_, e := colGroup.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id,
	})
	return e
}
func (s *Store) RemoveAllGroupMembers() error {
	_, e := colGroupMember.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id,
	})
	return e
}

// clear all things of the group
func (s *Store) find_group_id(gid string) (primitive.ObjectID, error) {
	g := &def.Group{}
	e := colGroup.FindOne(ctx, bson.M{
		`AccId`: s.acc_id, `Gid`: gid,
	}).Decode(g)
	if e != nil {
		return primitive.NilObjectID, e
	}
	return g.ID, nil
}

// clear all things of the group
func (s *Store) RemoveGroup(gid string) error {
	// 1. delete group
	id, e := s.find_group_id(gid)
	if e != nil {
		return e
	}

	_, e = colGroup.DeleteOne(ctx, bson.M{
		`AccId`: s.acc_id, `ID`: id,
	})
	if e != nil {
		return e
	}
	// 2. delete members
	_, e = colGroupMember.DeleteMany(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: id,
	})
	return e
}

// remove 1 member
func (s *Store) RemoveOneGroupMember(gid, jid string) error {
	id, e := s.find_group_id(gid)
	if e != nil {
		return e
	}
	_, e = colGroupMember.DeleteOne(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: id, `Jid`: jid,
	})
	return e
}

// add 1 member
func (s *Store) AddGroupMember(gid, jid string) error {
	id, e := s.find_group_id(gid)
	if e != nil {
		return e
	}
	_, e = colGroupMember.InsertOne(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: id, `Jid`: jid,
	})
	return e
}
func (s *Store) ListGroupMember(gid string, include_self bool) ([]*def.GroupMember, error) {
	if gid == `status@broadcast` { // sns
		return []*def.GroupMember{}, nil // TODO, log usync to db
	}
	id, e := s.find_group_id(gid)
	if e != nil {
		return nil, e
	}

	cur, e := colGroupMember.Find(ctx, bson.M{
		`AccId`: s.acc_id, `GroupId`: id,
	})
	if e != nil {
		return nil, e
	}
	var myJid string
	if include_self {
		myJid, e = s.GetMyJid()
		if e != nil {
			return nil, e
		}
	}
	members := []*def.GroupMember{}
	for cur.Next(ctx) {
		gm := &def.GroupMember{}
		e := cur.Decode(gm)
		if e != nil {
			return nil, e
		}

		if !include_self && gm.Jid == myJid {
		} else {
			members = append(members, gm)
		}
	}
	return members, nil
}
func (s *Store) ListGroupMemberJid(gid string, include_self bool) ([]string, error) {
	recs, e := s.ListGroupMember(gid, include_self)
	if e != nil {
		return nil, e
	}
	var jids []string
	for _, rec := range recs {
		jids = append(jids, rec.Jid)
	}

	return jids, nil
}
func (s *Store) GetWamSchedule() (*def.WamSchedule, error) {
	sch := &def.WamSchedule{}

	e := colWamSchedule.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(sch)

	if errors.Is(e, mongo.ErrNoDocuments) {
		_, e = colWamSchedule.InsertOne(ctx, bson.M{`AccId`: s.acc_id})
	}
	return sch, e
}
func (s *Store) ModifyWamSchedule(mod bson.M) error {
	r := colWamSchedule.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) WamSetNewNoiseLogin(isNewLogin bool) error {
	return s.ModifyWamSchedule(bson.M{
		`IsNewNoiseLogin`: isNewLogin,
	})
}
func (s *Store) WamSetHasNewMessage(hasNewMsg bool) error {
	return s.ModifyWamSchedule(bson.M{
		`HasNewMsg`: hasNewMsg,
	})
}

func (s *Store) GetWamEvent() (*def.WamEvent, error) {
	s.muWamEvent.RLock()
	defer s.muWamEvent.RUnlock()

	ret := &def.WamEvent{}

	e := colWamEvent.FindOne(ctx, bson.M{
		`AccId`: s.acc_id,
	}).Decode(ret)

	// must exists, inited when acc created
	return ret, e
}

func (s *Store) ModifyWamEvent(mod bson.M) error {
	s.muWamEvent.Lock()
	defer s.muWamEvent.Unlock()

	r := colWamEvent.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: mod,
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) AddWamEventBufs(
	evt_buf_arr [][]byte,
) error {
	s.muWamEvent.Lock()
	defer s.muWamEvent.Unlock()

	r := colWamEvent.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$push`: bson.M{
			`Buffer`: bson.M{
				`$each`: evt_buf_arr,
			},
		},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}

func (s *Store) ResetWamEventBuf() error {
	s.muWamEvent.Lock()
	defer s.muWamEvent.Unlock()

	r := colWamEvent.FindOneAndUpdate(ctx, bson.M{
		`AccId`: s.acc_id,
	}, bson.M{
		`$set`: bson.M{`Buffer`: [][]byte{}},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))
	return r.Err()
}
