package core

import (
	"fmt"
	"strconv"
	"time"

	"ajson"
	"aproto"
	"arand"
	"i"
	"wa/def"
	"wa/noise"
	"wa/pb"
	"wa/wam"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/proto"
)

// on(Ev_Noise_Location)
func New_Hook_NoiseLocation(a *Acc) func(...any) error {
	return func(args ...any) error {
		loc, _ := args[0].(string)
		return a.Store.SaveNoiseLocation(loc)
	}
}

func (c Core) Connect(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	// wam
	login_begin := time.Now()

	// prevent call Connect simultaneously
	a.Noise.MtxConnected.Lock()
	defer a.Noise.MtxConnected.Unlock()

	if a.Noise.IsConnected() {
		return NewErrRet(errors.New(`already connected`))
	}

	if j.Exists(`reset`) {
		e = a.Store.ModifyConfig(bson.M{
			`RemoteStatic`:    nil,
			`IsPassiveActive`: false,
		})

		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail reset`))
		}
	}

	prof, e := a.Store.GetProfile()
	if e != nil {
		return NewErrRet(e)
	}
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}

	full_phone, e := strconv.Atoi(dev.Cc + dev.Phone)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `invalid cc/phone`))
	}
	sessid := uint32(arand.Int(0, 0x7fffffff))

	var v1, v2, v3, v4 int32
	_, e = fmt.Sscanf(def.VERSION(dev.IsBusiness), "%d.%d.%d.%d", &v1, &v2, &v3, &v4)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `invalid version: `+def.VERSION(dev.IsBusiness)))
	}

	hsR_pub := cfg.RemoteStatic

	var isXX int32 = 1

	if len(hsR_pub) > 0 && cfg.IsPassiveActive {
		isXX = 0
	}

	pDev := &pb.NoiseHandshakeDevice{
		FullPhone: proto.Uint64(uint64(full_phone)),
		Passive:   proto.Int32(isXX), // 1: XX, 0: IK
		Struct_5: &pb.NoiseHandshakeDevice_XStruct_5{
			SMB_Android: proto.Int32(0),
			Ver: &pb.NoiseHandshakeDevice_XStruct_5_Version{
				V1: proto.Int32(v1),
				V2: proto.Int32(v2),
				V3: proto.Int32(v3),
				V4: proto.Int32(v4),
			},
			Mcc:            &dev.Mcc,
			Mnc:            &dev.Mnc,
			AndroidVersion: &dev.AndroidVersion,
			Brand:          &dev.Brand,
			Product:        &dev.Product,
			Build:          &dev.Build,
			Fdid:           &dev.Fdid,
			Language:       &dev.Language,
			Locale:         &dev.Locale,
			Board:          &dev.Board,
		},
		SessionId:      proto.Uint32(sessid),                 // pro 9
		Int_10:         proto.Int32(0),                       // pro 10
		NetworkSubType: proto.Int32(dev.NetSubType),          // pro 12
		SignatureMatch: proto.Int32(1),                       // pro 23
		ConnectionLc:   proto.Int32(int32(cfg.ConnectionLC)), // pro 24
	}
	if isXX == 0 && len(prof.Nick) > 0 {
		pDev.Nick = proto.String(prof.Nick)
	}
	if def.IsBeta(dev.IsBusiness) {
		pDev.Struct_5.IsBeta = proto.Int32(1)
		pDev.Dns = &pb.NoiseHandshakeDevice_DNS{
			Config: proto.Int32(0),
			Int_16: proto.Int32(0),
		}
		pDev.ConnectionSequenceAttempts = proto.Int32(1) // starts from 1
	}
	if dev.IsBusiness {
		// 10 for business, (both beta or not)
		pDev.Struct_5.SMB_Android = proto.Int32(10)
	}

	if len(hsR_pub) == 0 { // no RemoteStatic, use XX
		tmp, _ := proto.Marshal(pDev)
		a.Log.Info("handshake XX:\n %s", aproto.Dump(tmp))

		hsR_pub, e = a.Noise.HandshakeXX(
			noise.DHKey{Private: cfg.StaticPriv, Public: cfg.StaticPub},
			tmp, cfg.RoutingInfo)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `handshake`))
		}
		if e := a.Store.SaveRemoteNoiseStatic(hsR_pub); e != nil {
			return NewErrRet(errors.Wrap(e, `fail save hsR_pub`))
		}
	} else { // IK
		tmp, _ := proto.Marshal(pDev)
		a.Log.Info("handshake IK:\n%s", aproto.Dump(tmp))
		e = a.Noise.HandshakeIK(
			noise.DHKey{Private: cfg.StaticPriv, Public: cfg.StaticPub},
			hsR_pub,
			tmp, cfg.RoutingInfo)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `handshake, retry with param 'reset'`))
		}
	}

	// WamLogin
	if e := a.wam_login(login_begin, isXX); e != nil {
		a.Log.Error(`a.wam_login fail: ` + e.Error())
	}

	a.Fire(def.Ev_Connected, nil)

	// ConnectionLC ++ on success
	a.Store.ModifyConfig(bson.M{
		`ConnectionLc`: cfg.ConnectionLC + 1,
	})

	return NewSucc()
}

func (a *Acc) wam_login(
	login_begin time.Time, isXX int32,
) error {
	return a.AddWamEventBuf(wam.WamLogin, func(cc *wam.ClassChunk) {
		// 4: has noise client static key
		// 1: no noise client static key
		cc.Append(10, i.F(isXX == 1, 1, 4))

		loginT := int(time.Since(login_begin).Milliseconds())
		// with statistics, half percent is '3', other is '1'
		fifty_percent := arand.Int(0, 2) == 0
		if fifty_percent {
			cc.Append(6, 3)
			// has 'connectionT' when connectionOrigin == 3
			if loginT <= 1600 { // <= 1.6 sec
				cc.Append(5, int(loginT/2))
			} else {
				cc.Append(5, int(loginT)-arand.Int(200, 800))
			}
		} else {
			cc.Append(6, 1)
		}

		cc.Append(1, 1)
		cc.Append(3, loginT)
		cc.Append(4, 0)
		cc.Append(8, i.F(isXX == 1, 1, 0))
		cc.Append(2, 0)
		cc.Append(7, 0)
	})
}
