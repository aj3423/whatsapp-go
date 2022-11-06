package core

import (
	"encoding/json"
	"fmt"
	"time"

	"ahex"
	"ajson"
	"algo"
	"wa/def"
	"wa/def/clone"
	"wa/signal/ecc"
	"wa/signal/state/record"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (c Core) GetNoiseKey(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	cfg, e := a.Store.GetConfig()

	if e != nil {
		return NewErrRet(e)
	}

	rj := NewSucc()
	rj.Set(`StaticPriv`, algo.B64Enc(cfg.StaticPriv))
	rj.Set(`StaticPub`, algo.B64Enc(cfg.StaticPub))

	return rj
}

func (c Core) SetNoiseKey(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	StaticPriv, e := j.Get("StaticPriv").TryString()
	if e != nil {
		return NewErrRet(errors.New("invalid StaticPriv"))
	}

	priv_me, e := algo.B64Dec(StaticPriv)
	if e != nil {
		return NewErrRet(errors.New("invalid StaticPriv b64"))
	}

	x := ecc.CreateKeyPair(priv_me)
	pub := x.PublicKey().PublicKey()
	priv := x.PrivateKey().Serialize()

	e = a.Store.ModifyConfig(bson.M{
		`StaticPriv`: priv[:],
		`StaticPub`:  pub[:],

		// fixed
		`RemoteStatic`: ahex.Dec("a895af4adb4da29aa04360a05d84dce2399250a59dd51bf641331b4326292b06"),
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}

func (c Core) Clone2(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	c_b64, e := j.Get("c").TryString()
	if e != nil {
		return NewErrRet(errors.New("invalid 'c'"))
	}

	fmt.Println(c_b64)
	c_xml, e := algo.B64RawUrlDec(c_b64)
	if e != nil {
		c_xml, e = algo.B64Dec(c_b64)
		if e != nil {
			return NewErrRet(errors.New("invalid c b64"))
		}
	}

	var priv, pub []byte
	{
		doc := etree.NewDocument()
		if e := doc.ReadFromBytes(c_xml); e == nil {
			client_static_keypair_pwd_enc := doc.FindElements("//map/string[@name='client_static_keypair_pwd_enc']")[0].Text()
			j, e := ajson.Parse(client_static_keypair_pwd_enc)
			if e != nil {
				return NewErrRet(e)
			}

			staticI_enc, e1 := algo.B64PadDec(j.GetIndex(1).String())

			iv, e2 := algo.B64PadDec(j.GetIndex(2).String())

			salt, e3 := algo.B64PadDec(j.GetIndex(3).String())

			password := j.GetIndex(4).String()

			if e1 != nil || e2 != nil || e3 != nil {
				return NewErrRet(errors.New("parse error e1|e2|e3"))
			}
			key := algo.PbkdfSha1(
				append(def.RC2_FIXED_25, []byte(password)...),
				salt, 16, 16,
			)

			bs, e := algo.AesOfbDecrypt(staticI_enc, key, iv, &algo.None{})
			if e != nil {
				return NewErrRet(e)
			}
			priv = bs[0:0x20]
			pub = bs[0x20:]
		}
	}
	if len(priv) != 0x20 || len(pub) != 0x20 {
		return NewErrRet(errors.New("wrong pub priv"))
	}

	e = a.Store.ModifyConfig(bson.M{
		`StaticPriv`: priv[:],
		`StaticPub`:  pub[:],

		// fixed
		`RemoteStatic`: ahex.Dec("a895af4adb4da29aa04360a05d84dce2399250a59dd51bf641331b4326292b06"),
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}

func (c Core) PhoneClone(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	_ = a

	gene := &clone.Gene{}
	e = json.Unmarshal([]byte(j.Get("gene").ToString()), gene)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail parse gene`))
	}

	{
		mod := bson.M{}
		if gene.Nick != `` {
			mod[`Nick`] = gene.Nick
		}
		if len(mod) > 0 {
			if e = a.Store.ModifyProfile(mod); e != nil {
				return NewErrRet(e)
			}
		}
	}
	{
		mod := bson.M{}
		if len(gene.Fdid) > 0 {
			mod[`Fdid`] = gene.Fdid
		}
		if gene.RegistrationId != 0 {
			mod[`RegId`] = gene.RegistrationId
		}

		if len(mod) > 0 {
			if e = a.Store.ModifyDev(mod); e != nil {
				return NewErrRet(e)
			}
		}
	}
	{
		mod := bson.M{}

		if len(gene.RoutingInfo) > 0 {
			mod[`RoutingInfo`] = gene.RoutingInfo
		}
		if len(gene.NoiseLocation) > 0 {
			mod[`NoiseLocation`] = gene.NoiseLocation
		}
		if len(gene.AbPropsConfigKey) > 0 {
			mod[`AbPropsConfigKey`] = gene.AbPropsConfigKey
		}
		if len(gene.AbPropsConfigHash) > 0 {
			mod[`AbPropsConfigHash`] = gene.AbPropsConfigHash
		}
		if len(gene.ServerPropsConfigKey) > 0 {
			mod[`ServerPropsConfigKey`] = gene.ServerPropsConfigKey
		}
		if len(gene.ServerPropsConfigHash) > 0 {
			mod[`ServerPropsHash`] = gene.ServerPropsConfigHash
		}
		if len(gene.StaticPub) > 0 {
			mod[`StaticPub`] = gene.StaticPub
		}
		if len(gene.StaticPriv) > 0 {
			mod[`StaticPriv`] = gene.StaticPriv
		}

		if len(mod) > 0 {
			if e = a.Store.ModifyConfig(mod); e != nil {
				return NewErrRet(e)
			}
		}
	}
	{
		mod := bson.M{}
		if len(gene.PublicKey) > 0 {
			mod[`PublicKey`] = gene.PublicKey
		}
		if len(gene.PrivateKey) > 0 {
			mod[`PrivateKey`] = gene.PrivateKey
		}
		if gene.NextPrekeyId != 0 {
			mod[`NextPrekeyId`] = gene.NextPrekeyId
		}

		if len(mod) > 0 {
			if e = a.Store.ModifyMyIdentity(mod); e != nil {
				return NewErrRet(e)
			}
		}
	}
	{
		for _, pk := range gene.Prekeys {
			rec, e := record.NewPreKeyFromBytes(pk.Record)
			if e != nil {
				return NewErrRet(e)
			}
			if e = a.Store.StorePreKey(uint32(pk.PrekeyId), rec); e != nil {
				return NewErrRet(e)
			}
		}
	}
	{
		mod := bson.M{}
		if len(gene.Record) > 0 {
			mod[`PrekeyId`] = gene.PrekeyId
			mod[`Record`] = gene.Record
		}

		if len(mod) > 0 {
			if e = a.Store.ModifySignedPrekey(mod); e != nil {
				return NewErrRet(e)
			}
		}
	}

	// Schedule
	if !j.Exists(`full_init`) {
		now := time.Now()

		mod := bson.M{}
		mod[`GetPropsW`] = now
		mod[`GetPropsAbt`] = now
		mod[`CreateGoogle`] = now
		//mod[`ListGroup`] = now
		mod[`GetWbList`] = now
		mod[`GetProfilePicturePreview`] = now
		mod[`SetBackupToken`] = now
		mod[`SetEncrypt`] = now
		mod[`GetStatusPrivacy`] = now
		mod[`GetLinkedAccouts`] = now
		mod[`ThriftQueryCatkit`] = now
		mod[`GetBlockList`] = now
		mod[`UsyncDevice`] = now
		mod[`VerifyApps`] = now
		mod[`SetBizVerifiedName`] = now
		mod[`GetBizVerifiedName`] = now
		mod[`GetBizCatalog`] = now

		mod[`SetBizProfile`] = now
		mod[`GetBizProfile_4`] = now
		mod[`GetBizProfile_116`] = now

		mod[`GetStatusUser`] = now
		mod[`GetPrivacy`] = now
		mod[`GetProfilePicture`] = now
		mod[`GetJabberIqPrivacy`] = now
		mod[`Attestation`] = now
		mod[`AvailableNick`] = now
		mod[`BizBlockReason`] = now
		mod[`Wam`] = now

		if e = a.Store.ModifySchedule(mod); e != nil {
			return NewErrRet(e)
		}
	}
	// Wam
	if !j.Exists(`full_init`) {
		now := time.Now()

		mod := bson.M{}
		mod[`Basic`] = now
		mod[`WamLogin`] = now
		mod[`WamPsIdCreate`] = now
		mod[`WamDaily`] = now
		mod[`WamRegistrationComplete`] = now
		mod[`WamAndroidDatabaseMigrationDailyStatus`] = now
		mod[`WamStatusDaily`] = now
		mod[`WamAndroidDatabaseMigrationEvent`] = now
		mod[`WamSmbVnameCertHealth`] = now
		mod[`WamMessageSend`] = now
		mod[`WamE2eMessageSend`] = now
		mod[`WamMessageReceive`] = now

		if e = a.Store.ModifyWamSchedule(mod); e != nil {
			return NewErrRet(e)
		}
	}

	return NewSucc()
}
