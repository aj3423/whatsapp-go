package core

import (
	"time"

	"ajson"
	"arand"
	"wa/def"
	"wa/xmpp"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (a *Acc) get_disappearing_mode() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `disappearing_mode`},
		},
	})
	return e
}
func (a *Acc) get_block_list() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `blocklist`},
		},
	})
	return e
}
func (a *Acc) get_push_config() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `xmlns`, Value: `urn:xmpp:whatsapp:push`},
			{Key: `type`, Value: `get`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `config`,
				Attrs: []*xmpp.KeyValue{
					{Key: `version`, Value: `1`},
				},
			},
		},
	})
	return e
}
func (a *Acc) get_props(xmlns, protocol, hash string) (*xmpp.Node, error) {
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `xmlns`, Value: xmlns},
			{Key: `type`, Value: `get`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `props`,
				Attrs: []*xmpp.KeyValue{
					{Key: `protocol`, Value: protocol},
					{Key: `hash`, Value: hash},
				},
			},
		},
	})
	if e != nil {
		return nil, e
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return nil, e
	}
	mod := bson.M{}

	defer a.Store.ModifyConfig(mod)
	if xmlns == `w` {
		// save key/hash
		if ch, ok := nr.FindChildByTag(`props`); ok {
			attrs := ch.MapAttrs()
			if key, ok := attrs[`key`]; ok {
				mod[`ServerPropsConfigKey`] = key
			}
			if hash, ok := attrs[`hash`]; ok {
				mod[`ServerPropsHash`] = hash
			}
		}
	}
	if xmlns == `abt` {
		// save key/hash
		if ch, ok := nr.FindChildByTag(`props`); ok {
			attrs := ch.MapAttrs()
			if key, ok := attrs[`ab_key`]; ok {
				cfg.AbPropsConfigKey = key
			}
			if hash, ok := attrs[`hash`]; ok {
				cfg.AbPropsHash = hash
			}
		}
	}
	return nr, e
}
func (a *Acc) create_google() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `xmlns`, Value: `urn:xmpp:whatsapp:account`},
			{Key: `type`, Value: `get`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `crypto`,
				Attrs: []*xmpp.KeyValue{
					{Key: `action`, Value: `create`},
				},
				Children: []*xmpp.Node{
					{
						Tag:  `google`,
						Data: arand.Bytes(0x20),
					},
				},
			},
		},
	})
	return e
}
func (a *Acc) get_wb_list() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:b`},
			{Key: `to`, Value: `s.whatsapp.net`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `lists`,
			},
		},
	})
	return e
}
func (a *Acc) get_linked_accounts() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:biz`},
			{Key: `to`, Value: `s.whatsapp.net`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `linked_accounts`,
				Attrs: []*xmpp.KeyValue{
					{Key: `v`, Value: `3`},
				},
			},
		},
	})
	return e
}

// only in beta?
func (c Core) ThriftQueryCatkit(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `fb:thrift_iq`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `request`,
				Attrs: []*xmpp.KeyValue{
					{Key: `op`, Value: `typeahead`},
					{Key: `type`, Value: `catkit`},
					{Key: `v`, Value: `1`},
				},
				Children: []*xmpp.Node{
					{
						Tag:  `query`,
						Data: []byte{},
					},
				},
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}
	return NewJsonRet(nr.ToJson())
}

func (a *Acc) get_biz_block_reason() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:biz`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `mobile_config`,
				Attrs: []*xmpp.KeyValue{
					{Key: `name`, Value: `biz_block_reasons`},
					{Key: `v`, Value: `1`},
				},
			},
		},
	})
	return e
}
func (a *Acc) get_status_user() error {
	dev, e := a.Store.GetDev()
	if e != nil {
		return e
	}

	_, e = a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `status`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `status`,
				Children: []*xmpp.Node{
					{
						Tag: `user`,
						Attrs: []*xmpp.KeyValue{
							{Key: `jid`, Value: dev.Cc + dev.Phone + `@s.whatsapp.net`},
						}},
				},
			},
		},
	})
	return e
}
func (a *Acc) set_passive(active_state string) error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `passive`},
		},
		Children: []*xmpp.Node{
			{
				Tag: active_state, // `active`/`passive`
			},
		},
	})
	if e != nil {
		return e
	}
	e = a.Store.ModifyConfig(bson.M{
		`IsPassiveActive`: true,
	})
	if e != nil {
		return e
	}

	return nil
}

func (c Core) SetPassive(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	active_state, e := j.Get(`active_state`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param "active_state"`))
	}

	e = a.set_passive(active_state)
	if e != nil {
		return NewErrRet(e)
	}
	return NewSucc()
}
func (a *Acc) set_backup_token() error {
	dev, e := a.Store.GetDev()
	if e != nil {
		return e
	}

	_, e = a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:auth:backup:token`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `token`,
				/*
					phone uses:
						KeyGenerator v2 = KeyGenerator.getInstance("AES");
						v2.init(0xA0, SecureRandom.getInstance("SHA1PRNG"));
						return v2.generateKey().getEncoded();

					I just simply use dev.BackupToken
				*/
				Data: dev.BackupToken,
			},
		},
	})
	return e
}

func (c Core) Initialize(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	sch, e := a.Store.GetSchedule()
	if e != nil {
		return NewErrRet(e)
	}
	schMod := bson.M{}
	defer a.Store.ModifySchedule(schMod)

	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}

	//====================================
	{ // 1, every time
		e := a.get_push_config()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_push_config`))
		}
	}
	// 2. get_props_w. every day
	if sch.GetPropsW.Add(24 * time.Hour).Before(time.Now()) { // > 1 day
		_, e := a.get_props(`w`, `2`, cfg.ServerPropsHash)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_props_w`))
		}
		schMod[`GetPropsW`] = time.Now()
	}
	// 3. every day
	if sch.GetPropsAbt.Add(24 * time.Hour).Before(time.Now()) { // > 1 day
		_, e := a.get_props(`abt`, `1`, cfg.AbPropsHash)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_props_abt`))
		}
		schMod[`GetPropsAbt`] = time.Now()
	}
	{ // 4, every time
		e := a.presence(`available`, ``, ``)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail presence available`))
		}
	}
	// last time value is:
	//   7857dc757b6b0dfc13f35b0a95c7773b1a69b77804efc456e84dedf95d4e9a73
	if sch.CreateGoogle.IsZero() { // 5. once
		e := a.create_google()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail create_google`))
		}
		schMod[`CreateGoogle`] = time.Now()
	}

	if sch.ListGroup.IsZero() { // 6. once
		rj := c.ListGroup(j)
		if rj.Get(`ErrCode`).Int() != 0 {
			return rj
		}
		schMod[`ListGroup`] = time.Now()
	}

	if sch.GetWbList.IsZero() { // 7. once
		e := a.get_wb_list()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_wb_list`))
		}
		schMod[`GetWbList`] = time.Now()
	}

	if dev.IsBusiness { // once
		if sch.GetLinkedAccouts.IsZero() {
			e := a.get_linked_accounts()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get_linked_accounts`))
			}
			schMod[`GetLinkedAccouts`] = time.Now()
		}
	}
	if dev.IsBusiness { // once
		if sch.GetBizProfile_4.IsZero() {
			e := a.get_biz_profile(
				dev.Cc+dev.Phone+"@s.whatsapp.net",
				`4`, // catalog_status
			)
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get_business_profile`))
			}
			schMod[`GetBizProfile_4`] = time.Now()
		}
	}

	// 8. once
	if sch.GetProfilePicturePreview.IsZero() {
		j := ajson.New()
		j.Set(`type`, `preview`)
		j.Set(`jid`, dev.Cc+dev.Phone+`@s.whatsapp.net`)

		_, e := a.get_profile_picture(j)

		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_profile_picture`))
		}
		schMod[`GetProfilePicturePreview`] = time.Now()
	}

	// 9. once
	if sch.SetBackupToken.IsZero() {
		e := a.set_backup_token()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail set_backup_token`))
		}
		schMod[`SetBackupToken`] = time.Now()
	}

	// ignore 2fa ?

	if sch.SetEncrypt.IsZero() { // once
		e := a.set_encrypt()
		if e != nil {
			return NewErrRet(e)
		}
		schMod[`SetEncrypt`] = time.Now()
	}

	if !dev.IsBusiness { // 14. once
		if sch.GetStatusUser.IsZero() {
			e := a.get_status_user()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get_status_user`))
			}
			schMod[`GetStatusUser`] = time.Now()
		}
	}

	{ // every time
		e := a.set_media_conn()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail set_media_conn`))
		}
	}

	if sch.GetStatusPrivacy.IsZero() { // once
		e := a.get_privacy(`status`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `get status privacy`))
		}
		schMod[`GetStatusPrivacy`] = time.Now()
	}

	if sch.GetDisappearingMode.IsZero() { // once
		e := a.get_disappearing_mode()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_disappearing_mode`))
		}
		schMod[`GetDisappearingMode`] = time.Now()
	}

	if sch.GetPrivacy.IsZero() { // once
		e := a.get_privacy(`privacy`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_privacy`))
		}
		schMod[`GetPrivacy`] = time.Now()
	}

	if sch.GetProfilePicture.IsZero() { // once
		j := ajson.New()
		j.Set(`type`, `image`)
		j.Set(`query`, `url`)
		j.Set(`jid`, dev.Cc+dev.Phone+`@s.whatsapp.net`)

		_, e := a.get_profile_picture(j)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_profile_picture`))
		}
		schMod[`GetProfilePicture`] = time.Now()
	}

	if sch.GetBlockList.IsZero() { // once
		// send twice for personal
		{
			e := a.get_block_list()
			if e != nil {
				return NewErrRet(e)
			}
		}
		if !dev.IsBusiness { // twice
			e := a.get_block_list()
			if e != nil {
				return NewErrRet(e)
			}
		}
		schMod[`GetBlockList`] = time.Now()
	}

	// TODO google gcm

	if sch.UsyncDevice.IsZero() { // once
		_, e := a.usync_device(`query`, `notification`, `devices`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail usync_device`))
		}
		schMod[`UsyncDevice`] = time.Now()
	}

	// groups dirty
	// cfg may modified by hook, so get it here again
	{
		sch_, e := a.Store.GetSchedule()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail GetSchedule`))
		}
		if sch.IsGroupsDirty {
			e := a.set_dirty_clean(`groups`)
			if e != nil {
				return NewErrRet(e)
			}
			if e = a.Store.ModifySchedule(bson.M{
				`IsGroupsDirty`: false,
			}); e != nil {
				return NewErrRet(errors.Wrap(e, `fail group clean`))
			}
		}
		if sch_.IsAccountSyncDirty {
			e := a.set_dirty_clean(`account_sync`)
			if e != nil {
				return NewErrRet(e)
			}
			if e := a.Store.ModifySchedule(bson.M{
				`IsAccountSyncDirty`: false,
			}); e != nil {
				return NewErrRet(errors.Wrap(e, `fail set_dirty_clean`))
			}
		}
	}

	// TODO attestation "jws" ?

	{
		sch2, e := a.Store.GetSchedule()
		if e != nil {
			return NewErrRet(e)
		}
		if sch2.SafetynetAttestation {
			e := a.attestation(dev.HasGooglePlay) // response to prev attestation ?
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail attestation`))
			}
			schMod[`SafetynetAttestation`] = false
		}

		if !dev.IsBusiness {
			if sch2.SafetynetVerifyApps {
				e := a.verify_apps(dev.HasGooglePlay)
				if e != nil {
					return NewErrRet(errors.Wrap(e, `fail verify_apps`))
				}
				schMod[`SafetynetVerifyApps`] = false
			}
		}
	}

	if sch.AvailableNick.IsZero() { // 21. once
		prof, e := a.Store.GetProfile()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get nick`))
		}

		// input name
		delay_1 := arand.Int(10000, 20000)
		time.Sleep(time.Duration(delay_1) * time.Millisecond)

		// biz version delays more, for choosing category
		var delay_2 = 0
		if dev.IsBusiness {
			delay_2 = arand.Int(20000, 40000)
			time.Sleep(time.Duration(delay_2) * time.Millisecond)
		}

		e = a.presence(`available`, prof.Nick, ``)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail presence nick`))
		}
		schMod[`AvailableNick`] = time.Now()

		{
			e = a.Store.ModifyWamEvent(bson.M{
				`RegFillBizInfoScreen`: delay_1 + delay_2,
			})
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail Store RegFillBizInfoScreen`))
			}
		}
	}

	if contacts, e := j.Get(`contacts`).TryJsonArray(); e == nil {
		mode := `full`
		context := `registration`
		sid := `full`
		_, e = a.usync_contact(mode, context, sid, contacts)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail usync_contact`))
		}
	}

	if dev.IsBusiness { // once
		if sch.SetBizVerifiedName.IsZero() {
			e := a.set_biz_verified_name()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail set biz verified name`))
			}
			schMod[`SetBizVerifiedName`] = time.Now()
		}
	}

	if dev.IsBusiness { // once
		if sch.SetBizProfile.IsZero() {
			prof, e := a.Store.GetProfile()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get nick`))
			}
			tj := ajson.New()
			if len(prof.BizCategory) > 0 {
				tj.Set(`BizCategory`, prof.BizCategory)
			}
			if len(prof.BizDescription) > 0 {
				tj.Set(`BizDescription`, prof.BizDescription)
			}
			if len(prof.BizAddress) > 0 {
				tj.Set(`BizAddress`, prof.BizAddress)
			}

			e = a.set_biz_profile(tj)
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail set biz profile`))
			}
			schMod[`SetBizProfile`] = time.Now()
		}
	}

	if !cfg.IsPassiveActive { // once
		e := a.set_passive(`active`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail set_passive`))
		}
		a.Store.ModifyConfig(bson.M{
			`IsPassiveActive`: true,
		})
	}

	// call get_profile_picture for all groups/contact

	if sch.GetWbList.IsZero() { // once
		if sch.GetWbList.Add(3 * time.Minute).After(time.Now()) { // < 3 min
			e := a.get_wb_list()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get_wb_list 2`))
			}
			schMod[`GetWbList`] = time.Now()
		}
	}

	if dev.IsBusiness { // once
		if sch.GetBizProfile_116.IsZero() {
			e := a.get_biz_profile(
				dev.Cc+dev.Phone+"@s.whatsapp.net",
				`116`, // "biz_profile_options"
			)
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get biz profile 116`))
			}
			schMod[`GetBizProfile_116`] = time.Now()
		}
	}

	if dev.IsBusiness { // once
		if sch.GetBizVerifiedName.IsZero() {
			e := a.get_biz_verified_name()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get biz verified_name`))
			}
			schMod[`GetBizVerifiedName`] = time.Now()
		}
	}

	a.StartPingCron()

	if dev.IsBusiness && def.IsBeta(dev.IsBusiness) { // once
		if sch.ThriftQueryCatkit.IsZero() {
			rj := c.ThriftQueryCatkit(j)
			if rj.Get(`ErrCode`).Int() != 0 {
				return rj
			}
			schMod[`ThriftQueryCatkit`] = time.Now()
		}
	}

	// TODO get biz catalog
	//if dev.IsBusiness { // once
	//if sch.GetBizCatalog.IsZero() {
	//e := a.get_biz_catalog()
	//if e != nil {
	//return NewErrRet(errors.Wrap(e, `fail get biz verified_name`))
	//}
	//schMod[`GetBizCatalog`] = time.Now()
	//}
	//}

	// account_sync dirty
	if dev.IsBusiness && def.IsBeta(dev.IsBusiness) {
		if sch.BizBlockReason.Add(24 * time.Hour).Before(time.Now()) { // > 1 day
			e := a.get_biz_block_reason()
			if e != nil {
				return NewErrRet(errors.Wrap(e, `fail get_biz_block_reason`))
			}
			schMod[`BizBlockReason`] = time.Now()
		}
	}

	a.StartWamCron()
	a.StartDailyCron()
	a.StartMediaConnCron()

	return NewSucc()
}
