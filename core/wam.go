package core

import (
	"strconv"
	"time"

	"arand"
	"wa/crypto"
	"wa/def"
	"wa/wam"
	"wa/xmpp"

	"go.mongodb.org/mongo-driver/bson"
)

func wild_time() []byte {
	wc := &wam.WildChunk{}
	now := time.Now()
	wc.Append(47, now.Unix())
	b, _ := wc.ToBytes()
	return b
}
func with_wild_time(b []byte) [][]byte {
	bs := [][]byte{wild_time()}
	return append(bs, b)
}
func (a *Acc) AddWamEventBuf(
	cls *wam.ClassAttr,
	builder func(*wam.ClassChunk),
) error {
	if !cls.RateHit() {
		return nil
	}

	cc := &wam.ClassChunk{Id: cls.Id, Value: cls.Weight}
	builder(cc)
	b, e := cc.ToBytes()
	if e != nil {
		return e
	}

	var bs [][]byte

	// compare to last event time
	ev, e := a.Store.GetWamEvent()
	if e != nil {
		return e
	}
	now := time.Now()
	if ev.LastEventTime.Unix() != now.Unix() {
		bs = append(bs, wild_time())
	}
	bs = append(bs, b)

	// add event buf
	e = a.Store.AddWamEventBufs(bs)
	if e != nil {
		return e
	}
	// update last time
	if e := a.Store.ModifyWamEvent(bson.M{
		`LastEventTime`: now,
	}); e != nil {
		a.Log.Error(`fail ModifyWamEvent: ` + e.Error())
	}

	return nil
}

func year_class(dev *def.Device) int32 {
	if dev.YearClass != 0 {
		return dev.YearClass
	}
	switch dev.AndroidApiLevel {
	case 26: // Android 7
		return 2013
	default:
		return 2014
	}
}
func year_class_2016(dev *def.Device) int32 {
	if dev.YearClass2016 != 0 {
		return dev.YearClass2016
	}
	return 2016
}
func wam_basic_block(dev *def.Device, cfg *def.Config) ([]byte, error) {
	wc := wam.WildChunk{}
	wc.Append(11, 2)
	wc.Append(7335, arand.Int(0, 2)) // 0/1
	wc.Append(13, dev.Brand+"-"+dev.Model)
	wc.Append(2795, cfg.NoiseLocation)
	wc.Append(1657, 4)
	if i, e := strconv.Atoi(dev.Mcc); e != nil {
		wc.Append(5, i)
	}
	wc.Append(5029, "") // TODO ab_props:sys:last_exposure_keys
	if i, e := strconv.Atoi(dev.Mnc); e != nil {
		wc.Append(3, i)
	}
	wc.Append(1659, 1)
	wc.Append(15, dev.AndroidVersion)
	wc.Append(495, dev.Product)
	wc.Append(6251, 1)
	wc.Append(105, dev.NetSubType)
	wc.Append(23, dev.NetType)
	wc.Append(2617, year_class_2016(dev))
	wc.Append(655, dev.MemoryClass)
	wc.Append(4473, cfg.AbPropsConfigKey)
	wc.Append(17, def.VERSION(dev.IsBusiness))
	wc.Append(287, dev.Brand)
	wc.Append(289, dev.Model)
	wc.Append(21, 0)
	wc.Append(689, year_class(dev))
	//wc.Append(2141, cfg.ServerPropsConfigKey)

	b, e := wc.ToBytes()
	if e != nil {
		return nil, e
	}
	return b, nil
}

func (a *Acc) wam_stats() error {
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return e
	}
	dev, e := a.Store.GetDev()
	if e != nil {
		return e
	}
	wamEvt, e := a.Store.GetWamEvent()
	if e != nil {
		return e
	}

	// nothing to send
	if len(wamEvt.Buffer) == 0 {
		return nil
	}

	bs := wam.Head // 0x57, 0x41, 0x4d, 0x05
	bs = append(bs, 1)
	bs = append(bs, crypto.U162LE(wamEvt.ReqId+1)...)
	bs = append(bs, 0)

	basic, e := wam_basic_block(dev, cfg)
	if e != nil {
		return e
	}
	bs = append(bs, basic...)

	for _, b := range wamEvt.Buffer {
		bs = append(bs, b...)
	}
	/*
		editBusinessProfileSessionId := arand.Uuid4()
		if dev.IsBusiness && isFirstRegister {
			cc := wam.ClassChunk{Id: 1466, Value: 1}
			cc.Append(10, 1)
			cc.Append(2, editBusinessProfileSessionId)
			cc.Append(1, 1)
			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)
		}
		if dev.IsBusiness && isFirstRegister {
			cc := wam.ClassChunk{Id: 1466, Value: 1}
			cc.Append(10, 1)
			cc.Append(2, editBusinessProfileSessionId)
			cc.Append(1, 2)
			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)
		}
		// WamRegistrationComplete

		if dev.IsBusiness && isFirstRegister {
			cc := wam.ClassChunk{Id: 1578, Value: 1}
			cc.Append(2, 1)
			cc.Append(1, 9)
			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)
		}
		if dev.IsBusiness && isFirstRegister {
			cc := wam.ClassChunk{Id: 1578, Value: 1}
			cc.Append(2, 1)
			cc.Append(1, 9)
			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)
		}
		if dev.IsBusiness && isFirstRegister {
			cc := wam.ClassChunk{Id: 2222, Value: 1}
			cc.Append(1, 0)
			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)
		}

		// WamDaily

		// WamAndroidDatabaseMigrationDailyStatus
		if dev.IsBusiness && wamSch.WamAndroidDatabaseMigrationDailyStatus.Add(
			24*time.Hour).Before(time.Now()) { // > 1 day

			cc := wam.ClassChunk{Id: 2318, Value: -1}

			cc.Append(1, 1)
			cc.Append(7, 1)
			cc.Append(29, 1)
			cc.Append(4, 1)
			cc.Append(28, 1)
			cc.Append(27, 1)
			cc.Append(19, 4)
			cc.Append(3, 1)
			cc.Append(14, 4)
			cc.Append(6, 1)
			cc.Append(5, 1)
			cc.Append(10, 4)
			cc.Append(32, 4)
			cc.Append(11, 4)
			cc.Append(20, 4)
			cc.Append(25, 4)
			cc.Append(17, 4)
			cc.Append(2, 1)
			cc.Append(30, 1)
			cc.Append(24, 4)
			cc.Append(22, 4)
			cc.Append(15, 4)
			cc.Append(31, 1)
			cc.Append(33, 4)
			cc.Append(8, 1)
			cc.Append(9, 1)
			cc.Append(18, 4)
			cc.Append(23, 4)
			cc.Append(16, 4)
			cc.Append(12, 4)
			cc.Append(21, 4)
			cc.Append(13, 4)
			cc.Append(26, -1)

			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)

			modWamSch[`WamAndroidDatabaseMigrationDailyStatus`] = time.Now()
		}

		// WamStatusDaily
		if dev.IsBusiness && wamSch.WamStatusDaily.Add(
			24*time.Hour).Before(time.Now()) { // > 1 day

			cc := wam.ClassChunk{Id: 1676, Value: 1}

			cc.Append(3, 0)
			cc.Append(1, 0)
			cc.Append(4, 0)
			cc.Append(2, 0)

			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)

			modWamSch[`WamStatusDaily`] = time.Now()
		}

		// WamAndroidDatabaseMigrationEvent
		if dev.IsBusiness && wamSch.HasNewMsg && wamSch.WamAndroidDatabaseMigrationEvent.Add(
			24*time.Hour).Before(time.Now()) { // > 1 day

			cc := wam.ClassChunk{Id: 1912, Value: 1}

			sz := 925696
			sz += 64 * arand.Int(-10, 10)
			cc.Append(5, sz)
			cc.Append(4, sz)
			cc.Append(9, int64(dev.StorageAvailSize))
			cc.Append(1, `message_main`)
			cc.Append(10, 3)
			cc.Append(2, 1)
			cc.Append(3, 0)
			cc.Append(6, 0)
			cc.Append(7, 0)
			cc.Append(8, 0)

			b, e := cc.ToBytes()
			if e != nil {
				return e
			}
			bs = append(bs, b...)

			modWamSch[`hasNewMsg`] = false
			modWamSch[`WamAndroidDatabaseMigrationEvent`] = time.Now()
		}

		if dev.IsBusiness && isFirstRegister {
			{
				cc := wam.ClassChunk{Id: 1578, Value: 1}
				cc.Append(2, 1)
				cc.Append(1, 9)
				b, e := cc.ToBytes()
				if e != nil {
					return e
				}
				bs = append(bs, b...)
			}
			{
				cc := wam.ClassChunk{Id: 1578, Value: 1}
				cc.Append(2, 1)
				cc.Append(1, 9)
				b, e := cc.ToBytes()
				if e != nil {
					return e
				}
				bs = append(bs, b...)
			}
			{
				t = t.Add(time.Duration(arand.Int(1, 3)) * time.Second)
				bs = append(bs, wild_time(t)...)
			}
			{
				cc := wam.ClassChunk{Id: 2222, Value: 1}
				cc.Append(1, 2)
				b, e := cc.ToBytes()
				if e != nil {
					return e
				}
				bs = append(bs, b...)
			}
			{
				cc := wam.ClassChunk{Id: 1578, Value: 1}
				cc.Append(2, 1)
				cc.Append(1, 9)
				b, e := cc.ToBytes()
				if e != nil {
					return e
				}
				bs = append(bs, b...)
			}
			{
				t = t.Add(time.Duration(arand.Int(6, 12)) * time.Second)
				bs = append(bs, wild_time(t)...)
			}
			{
				cc := wam.ClassChunk{Id: 1578, Value: 1}
				cc.Append(2, 3)
				cc.Append(1, 9)
				b, e := cc.ToBytes()
				if e != nil {
					return e
				}
				bs = append(bs, b...)
			}
		}

		// WamSmbVnameCertHealth
		if dev.IsBusiness && wamSch.WamSmbVnameCertHealth.IsZero() { // once
			if 0 == arand.Int(0, 2) {
				{
					t = t.Add(time.Duration(arand.Int(5, 14)) * time.Second)
					bs = append(bs, wild_time(t)...)
				}
				{
					cc := wam.ClassChunk{Id: 1602, Value: 1}
					cc.Append(1, 1)
					b, e := cc.ToBytes()
					if e != nil {
						return e
					}
					bs = append(bs, b...)
					modWamSch[`WamSmbVnameCertHealth`] = time.Now()
				}
			}
		}
		// WamMessageSend
		// WamE2eMessageSend
		// WamMessageReceive
	*/
	_, e = a.Noise.WriteReadXmppNode(&xmpp.Node{
		Compressed: true,

		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:stats`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `add`,
				Attrs: []*xmpp.KeyValue{
					{Key: `t`, Value: strconv.Itoa(int(time.Now().Unix()))},
				},
				Data: bs,
			},
		},
	})
	if e != nil {
		return e
	}

	// save to db
	// update wamSch only if all success
	//a.Store.ModifyWamSchedule(modWamSch)
	a.Store.ModifyWamEvent(bson.M{
		`ReqId`: wamEvt.ReqId + 1,
	})

	// clear table WamEvent.Buffer
	a.Store.ResetWamEventBuf()

	return nil
}
