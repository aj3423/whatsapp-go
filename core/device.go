package core

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"

	"ahex"
	"ajson"
	"algo"
	"android/ver"
	"wa/def"
	"wa/signal/ecc"
	"wa/signal/keys/identity"
	"wa/signal/state/record"
	"wa/signal/util/bytehelper"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (c Core) SetDevice(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	e = a.Store.SetJsonDev(j)
	if e != nil {
		return NewErrRet(e)
	}

	return NewRet(0)
}

func (c Core) DumpDevice(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	bs, e := json.Marshal(dev)
	if e != nil {
		return NewErrRet(e)
	}

	devj, e := ajson.ParseByte(bs)
	if e != nil {
		return NewErrRet(e)
	}

	ret := NewRet(0)
	ret.Set("dev", devj)

	return ret
}

/* return
1. cc
2. phone
3. cc+phone.last6()
4. error
*/
func CheckNum(cc_phone string) (string, string, string, error) {
	reg := regexp.MustCompile("^([17]|2[07]|3[0123469]|4[013456789]|5[12345678]|6[0123456]|8[1246]|9[0123458]|\\d{3})\\d*?(\\d{4,6})$")
	m := reg.FindStringSubmatch(cc_phone)
	if len(m) != 3 {
		return ``, ``, ``, errors.New(`fail CheckNum, invalid cc/phone`)
	}
	cc := m[1]
	phone := cc_phone[len(cc):]
	return cc, phone, m[1] + m[2], nil
}

func (c Core) SetInterceptedDevice(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	var cc, phone string
	{ // me
		me, e := algo.B64RawUrlDec(j.Get(`b`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'b'`))
		}

		p := bytes.Index(me, []byte(`umbe`))
		p += 16 + 4
		rest := me[p:]

		var cc_phone string
		for len(rest) > 0 {
			b := rest[0]
			if b >= '0' && b <= '9' {
				cc_phone += string(rune(b))
			} else {
				break
			}
			rest = rest[1:]
		}

		cc, phone, _, e = CheckNum(cc_phone)
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'b'`))
		}

		e = a.Store.ModifyDev(bson.M{
			`Cc`:    cc,
			`Phone`: phone,
		})
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail save 'b'`))
		}
	}
	{ // rc2
		rc2, e := algo.B64RawUrlDec(j.Get(`a`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}

		rest := rc2
		p := bytes.Index(rest, []byte{0x2a, 0x00, 0x02})
		if p == -1 {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}
		rest = rest[p+3:]
		if len(rest) < 4 {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}
		salt := rest[0:4]
		rest = rest[4:]
		if len(rest) < 16 {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}
		iv := rest[0:16]
		rest = rest[16:]
		if len(rest) < 20 {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}
		enc := rest[0:20]

		_, _, chnum, e := CheckNum(cc + phone)
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}
		key := append(def.RC2_FIXED_25, []byte(chnum)...)
		k := algo.PbkdfSha1(key, salt, 16, 16)
		backup_token, e := algo.AesOfbDecrypt(enc, k, iv, &algo.None{})
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'a'`))
		}

		e = a.Store.ModifyDev(bson.M{
			`BackupToken`: backup_token,
		})
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail save 'a'`))
		}
	}

	{ // d,e: identity pub/priv
		iden_pub, e := algo.B64RawUrlDec(j.Get(`d`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'd'`))
		}
		idpub := ahex.Dec(string(iden_pub))
		if len(idpub) != 33 || idpub[0] != 0x05 {
			return NewErrRet(errors.New(`fail decode 'd'`))
		}
		idpub = idpub[1:]

		// e: identity_priv
		iden_priv, e := algo.B64RawUrlDec(j.Get(`e`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'e'`))
		}
		idpriv := ahex.Dec(string(iden_priv))
		if len(idpriv) != 32 {
			return NewErrRet(errors.New(`fail decode 'e'`))
		}

		iden := identity.NewKeyPair(
			identity.NewKeyFromBytes(bytehelper.SliceToArray(idpub)),
			ecc.NewDjbECPrivateKey(bytehelper.SliceToArray(idpriv)),
		)
		e = a.Store.SetMyIdentityKeyPair(iden)
		if e != nil {
			return NewErrRet(errors.New(`fail save 'd,e'`))
		}
	}
	{ // g: SignedPrekeyRecord
		spk_str, e := algo.B64RawUrlDec(j.Get(`g`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'g'`))
		}
		spk_pb := ahex.Dec(string(spk_str))
		spk_rec, e := record.NewSignedPreKeyFromBytes(spk_pb)
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'g'`))
		}
		e = a.Store.StoreSignedPreKey(spk_rec.ID(), spk_rec)
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'g'`))
		}
	}
	{ // aa~ah:
		m := bson.M{
			`Locale`:     j.Get(`aa`).String(),
			`Language`:   `en`,
			`NetType`:    1,
			`NetSubType`: 1,
			`HasSdCard`:  1,
			`CpuAbi`:     j.Get(`m`).String(),
			`Model`:      j.Get(`ab`).String(),
			`Brand`:      j.Get(`ac`).String(),
			// ad:  WTF?
			`Fdid`:           j.Get(`ae`).String(),
			`AndroidVersion`: j.Get(`af`).String(),
			`Product`:        j.Get(`h`).String(),
			`Board`:          j.Get(`i`).String(),
			`Build`:          j.Get(`j`).String(),
		}
		lvl := ver.ApiLevel(j.Get(`af`).String())
		if lvl == -1 {
			m[`AndroidApiLevel`] = 24 // default 24
		} else {
			m[`AndroidApiLevel`] = int32(lvl)
		}
		mcc := j.Get(`ag`).String()
		m[`Mcc`] = mcc
		if mcc == `` {
			m[`Mcc`] = `000`
		}
		mnc := j.Get(`ah`).String()
		m[`Mnc`] = mnc
		if mnc == `` {
			m[`Mnc`] = `000`
		}
		sz, e := strconv.Atoi(j.Get(`k`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'k'`))
		}
		m[`StorageAvailSize`] = int64(sz)
		m[`ExternalStorageAvailSize`] = int64(sz)

		sz, e = strconv.Atoi(j.Get(`l`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'l'`))
		}
		m[`StorageTotalSize`] = int64(sz)
		m[`ExternalStorageTotalSize`] = int64(sz)

		reg_id, e := strconv.Atoi(j.Get(`f`).String())
		if e != nil {
			return NewErrRet(errors.New(`fail decode 'f'`))
		}
		m[`RegId`] = uint32(reg_id)

		e = a.Store.ModifyDev(m)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail save 'aa~ah'`))
		}
	}
	{ // n: nick
		if nick, e := j.Get(`n`).TryString(); e == nil {
			e = a.Store.ModifyProfile(bson.M{
				`Nick`: nick,
			})
			if e != nil {
				return NewErrRet(e)
			}
		}
	}

	return NewRet(0)
}
