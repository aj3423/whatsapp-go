package core

import (
	"time"

	"ajson"
	"arand"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (c Core) InitializeEmu(j *ajson.Json) *ajson.Json {
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

	if sch.GetStatusPrivacy.IsZero() { // once
		e := a.get_privacy(`status`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `get status privacy`))
		}
		schMod[`GetStatusPrivacy`] = time.Now()
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
	{ // every time
		e := a.set_media_conn()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail set_media_conn`))
		}
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

	// groups dirty
	// cfg may modified by hook, so get it here again
	{
		sch_, e := a.Store.GetSchedule()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail GetSchedule`))
		}
		if sch_.IsGroupsDirty {
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

	if sch.UsyncDevice.IsZero() { // once
		_, e := a.usync_device(`query`, `notification`, `devices`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail usync_device`))
		}
		schMod[`UsyncDevice`] = time.Now()
	}

	if sch.AvailableNick.IsZero() { // 21. once
		prof, e := a.Store.GetProfile()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get nick`))
		}
		time.Sleep(time.Duration(arand.Int(5000, 15000)) * time.Millisecond)
		// biz version delays more, for choosing category
		if dev.IsBusiness {
			time.Sleep(time.Duration(arand.Int(5000, 12000)) * time.Millisecond)
		}

		e = a.presence(`available`, prof.Nick, ``)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail presence nick`))
		}
		schMod[`AvailableNick`] = time.Now()
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

	if sch.GetWbList.IsZero() { // once
		e := a.get_wb_list()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_wb_list 2`))
		}
		schMod[`GetWbList`] = time.Now()
	}

	if sch.GetPrivacy.IsZero() { // once
		e := a.get_privacy(`privacy`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail get_privacy`))
		}
		schMod[`GetPrivacy`] = time.Now()
	}
	if sch.GetStatusPrivacy.IsZero() { // once
		e := a.get_privacy(`status`)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `get status privacy`))
		}
		schMod[`GetStatusPrivacy`] = time.Now()
	}

	a.StartPingCron()

	return NewSucc()
}
