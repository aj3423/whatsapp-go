package core

import (
	"strconv"
	"time"

	"ahex"
	"algo"
	"arand"
	"wa/def"
	"wa/wam"

	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
)

func (a *Acc) StartPingCron() {
	if a.cronPing == nil {
		a.Wg.Add(1)
		a.cronPing = gocron.NewScheduler(time.UTC)

		// heart beat every 4 min
		a.cronPing.Every(4).Minutes().Do(func() {
			a.Event.Fire(def.Ev_Heartbeat)
		})
		a.cronPing.StartAsync()
	}
}
func (a *Acc) StopPingCron() {
	if a.cronPing != nil {
		a.cronPing.Stop()
		a.cronPing.Clear()
		a.cronPing = nil
		a.Wg.Done()
	}
}

func (a *Acc) StartDailyCron() {
	if a.cronDaily == nil {
		a.Wg.Add(1)
		a.cronDaily = gocron.NewScheduler(time.UTC)

		a.cronDaily.Every(5).Minutes().
			// delay first run 100 ~ 500ms
			StartAt(time.Now().Add(time.Duration(
				arand.Int(100, 500)) * time.Millisecond)).
			Do(func() {

				wamEvt, e := a.Store.GetWamEvent()
				if e != nil {
					a.Log.Error("fail GetWamEvent in cronDaily: " + e.Error())
					return
				}
				wamSch, e := a.Store.GetWamSchedule()
				if e != nil {
					a.Log.Error("fail GetWamSchedule in cronDaily: " + e.Error())
					return
				}
				dev, e := a.Store.GetDev()
				if e != nil {
					a.Log.Error("fail GetDev in cronDaily: " + e.Error())
					return
				}

				// WamRegistrationComplete
				if wamSch.WamRegistrationComplete.IsZero() {
					e := a.AddWamEventBuf(wam.WamRegistrationComplete, func(cc *wam.ClassChunk) {
						cc.Append(9, algo.B64RawUrlEnc(dev.ExpId))
						cc.Append(4, 0)
						cc.Append(7, 1)
						if dev.HasGooglePlay {
							cc.Append(10, 1)
						}
						cc.Append(5, 1)
						cc.Append(6, 1)
						cc.Append(3, 0)
						cc.Append(8, 1)

						// both personal/biz
						BizScreenT := wamEvt.RegFillBizInfoScreen
						registrationT := time.Since(wamEvt.EulaAccept).Milliseconds()
						cc.Append(1, registrationT)
						cc.Append(2, BizScreenT)
					})
					if e != nil {
						a.Log.Error(`fail WamRegistrationComplete : ` + e.Error())
					}

					if e := a.Store.ModifyWamSchedule(bson.M{
						`WamRegistrationComplete`: time.Now(),
					}); e != nil {
						a.Log.Error(`fail Modify WamRegistrationComplete schedule: ` + e.Error())
					}
				}

				// WamDaily
				if time.Since(wamSch.WamDaily) > 24*time.Hour {
					e := a.AddWamEventBuf(wam.WamDaily, func(cc *wam.ClassChunk) {
						sz_all := wamEvt.AddressBookSize
						sz_wa := wamEvt.AddressBookWASize
						if sz_wa == 0 {
							sz_wa = int32(arand.Int(1, 20))
							sz_all = sz_wa * int32(arand.Int(3, 10))

							a.Store.ModifyWamEvent(bson.M{
								`AddressBookSize`:   sz_all,
								`AddressBookWASize`: sz_wa,
							})
						}
						cc.Append(11, sz_all) // TODO addressbookSize == (3~10)*below
						cc.Append(12, sz_wa)  // TODO addressbookWhatsappSize == random(1, 20)
						cc.Append(37, dev.AndroidApiLevel)
						cc.Append(39, dev.HasSdCard)
						cc.Append(42, 1)
						cc.Append(41, 1)
						cc.Append(40, dev.IsSdCardRemovable)
						cc.Append(139, 3)
						cc.Append(98, 0)
						cc.Append(49, 0)
						cc.Append(103, ahex.Enc(def.AppCodeHash(dev.IsBusiness)))
						cc.Append(121, 5)
						cc.Append(48, 0)
						cc.Append(90, 0)
						cc.Append(91, 0)
						cc.Append(89, 1)
						cc.Append(96, 0)
						cc.Append(97, 0)
						cc.Append(95, 1)
						cc.Append(87, 1)
						cc.Append(88, 0)
						cc.Append(86, 1)
						cc.Append(93, 0)
						cc.Append(94, 0)
						cc.Append(92, 1)
						cc.Append(126, 0)
						cc.Append(138, 0)
						cc.Append(9, 0)
						cc.Append(128, 1)
						if !dev.IsBusiness {
							//cc.Append(18, 0)
							//cc.Append(17, 0)
						}
						db_sz := wamEvt.ChatDatabaseSize
						if db_sz == 0 {
							db_sz = int32(arand.Int(900000, 2000000) / 1024 * 1024)
							a.Store.ModifyWamEvent(bson.M{
								`ChatDatabaseSize`: db_sz,
							})
						}
						db_sz = db_sz + int32(arand.Int(-10, 10)*1024) // +- 10k
						cc.Append(19, db_sz)                           // chatDatabaseSize
						cc.Append(85, dev.CpuAbi)
						if arand.Bool() {
							cc.Append(140, 0)
						}
						cc.Append(153, dev.Language)
						cc.Append(109, dev.ExternalStorageAvailSize)
						cc.Append(110, dev.ExternalStorageTotalSize)
						cc.Append(112, 0)
						cc.Append(111, 0)
						cc.Append(119, 1)
						cc.Append(62, 0)
						if dev.HasGooglePlay {
							cc.Append(43, 1) // googlePlayServicesAvailable
							v, e := strconv.Atoi(dev.GooglePlayServiceVersion)
							if e == nil {
								cc.Append(79, v) // googlePlayServicesVersion
							} else {
								a.Log.Error("wrong device GooglePlayServiceVersion: " + e.Error())
							}
						} else {
							cc.Append(43, 0)  // googlePlayServicesAvailable
							cc.Append(79, -1) // googlePlayServicesVersion
						}

						if dev.HasGooglePlay {
							cc.Append(120, "com.android.vending")
						}
						cc.Append(137, 0)
						if dev.IsBusiness {
						} else {
							//cc.Append(16, 0) // TODO groupArchivedChatCount
							//grp_cnt, _ := a.Store.GroupCount()
							//cc.Append(15, grp_cnt)
							//cc.Append(14, 0) // TODO individualArchivedChatCount
							//cc.Append(13, 0) // TODO individualChatCount
						}
						cc.Append(115, 0)
						cc.Append(114, 0)

						//if !dev.IsBusiness {
						//cc.Append(45, 0)
						//}

						cc.Append(46, 0)
						cc.Append(60, 0)
						cc.Append(61, 0)
						cc.Append(38, 0)
						cc.Append(154, "")
						if dev.IsBusiness {
							cc.Append(82, 0)
							cc.Append(84, 0)
							cc.Append(83, 0)
						}

						cc.Append(5, dev.Language)
						cc.Append(44, 0)
						//if !dev.IsBusiness {
						//cc.Append(81, 0)
						//cc.Append(80, 0)
						//}
						cc.Append(6, dev.Locale)

						// TODO according to msg count
						// /sdcard/Media
						media_count := arand.Int(0, 30)
						cc.Append(21, media_count)
						cc.Append(20, media_count*3108) // 1 file == 3108 bytes in average

						cc.Append(155, 0)
						cc.Append(7, 0)
						cc.Append(4, dev.Build)
						cc.Append(118, 1)
						cc.Append(102, def.PKG_NAME(dev.IsBusiness))
						cc.Append(100, 0)
						cc.Append(57, -1)
						cc.Append(58, -1)
						cc.Append(56, 0)
						cc.Append(52, 0)
						cc.Append(50, 0)
						cc.Append(53, 0)
						cc.Append(59, 0)
						cc.Append(55, 0)
						cc.Append(51, 0)
						cc.Append(54, 0)
						cc.Append(156, dev.CpuCount)
						cc.Append(8, 1)
						cc.Append(77, def.SignatureHash)

						avail_size := int(dev.StorageAvailSize) + arand.Int(-1024, 1024)*1024
						cc.Append(31, avail_size)
						cc.Append(32, dev.StorageTotalSize)
						if wamSch.WamDaily.IsZero() {
							cc.Append(127, 0)
						} else {
							cc.Append(127, time.Since(wamSch.WamDaily).Milliseconds())
						}
						// TODO according to media(video) count
						// /sdcard/Media
						//video_count := media_count / 5
						cc.Append(23, 0)
						cc.Append(22, 0)
					})
					if e != nil {
						a.Log.Error(`fail WamDaily: ` + e.Error())
					}

					if e := a.Store.ModifyWamSchedule(bson.M{
						`WamDaily`: time.Now(),
					}); e != nil {
						a.Log.Error(`fail Modify WamDaily schedule: ` + e.Error())
					}
				}
				// WamPttDaily
				if time.Since(wamSch.WamPttDaily) > 24*time.Hour {
					e := a.AddWamEventBuf(wam.WamPttDaily, func(cc *wam.ClassChunk) {
						cc.Append(9, 0)
						cc.Append(8, 0)
						cc.Append(7, 0)
						cc.Append(15, 0)
						cc.Append(14, 0)
						cc.Append(13, 0)
						cc.Append(21, 0)
						cc.Append(20, 0)
						cc.Append(19, 0)
						cc.Append(12, 0)
						cc.Append(11, 0)
						cc.Append(10, 0)
						cc.Append(18, 0)
						cc.Append(17, 0)
						cc.Append(16, 0)
						cc.Append(3, 0)
						cc.Append(2, 0)
						cc.Append(1, 0)
						cc.Append(6, 0)
						cc.Append(5, 0)
						cc.Append(4, 0)
						cc.Append(25, 0)
						cc.Append(26, 0)
						cc.Append(27, 0)
					})
					if e != nil {
						a.Log.Error(`fail WamPttDaily: ` + e.Error())
					}

					if e := a.Store.ModifyWamSchedule(bson.M{
						`WamPttDaily`: time.Now(),
					}); e != nil {
						a.Log.Error(`fail Modify WamPttDaily schedule: ` + e.Error())
					}
				}
				// WamAndroidDatabaseMigrationDailyStatus
				if time.Since(wamSch.WamAndroidDatabaseMigrationDailyStatus) > 24*time.Hour {
					e := a.AddWamEventBuf(wam.WamAndroidDatabaseMigrationDailyStatus, func(cc *wam.ClassChunk) {
						cc.Append(1, 1)
						cc.Append(7, 1)
						cc.Append(29, 1)
						cc.Append(4, 1)
						cc.Append(36, 4)
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
						cc.Append(33, 5)
						cc.Append(8, 1)
						cc.Append(9, 1)
						cc.Append(35, 4)
						cc.Append(18, 4)
						cc.Append(23, 4)
						cc.Append(16, 4)
						cc.Append(12, 4)
						cc.Append(21, 4)
						cc.Append(13, 4)
						cc.Append(26, -1)
					})
					if e != nil {
						a.Log.Error(`fail WamAndroidDatabaseMigrationDailyStatus: ` + e.Error())
					}

					if e := a.Store.ModifyWamSchedule(bson.M{
						`WamAndroidDatabaseMigrationDailyStatus`: time.Now(),
					}); e != nil {
						a.Log.Error(`fail Modify WamAndroidDatabaseMigrationDailyStatus schedule: ` + e.Error())
					}
				}
				// WamStatusDaily
				if time.Since(wamSch.WamStatusDaily) > 24*time.Hour {
					e := a.AddWamEventBuf(wam.WamStatusDaily, func(cc *wam.ClassChunk) {
						cc.Append(3, 0)
						cc.Append(1, 0)
						cc.Append(4, 0)
						cc.Append(2, 0)
					})
					if e != nil {
						a.Log.Error(`fail WamStatusDaily: ` + e.Error())
					}

					if e := a.Store.ModifyWamSchedule(bson.M{
						`WamStatusDaily`: time.Now(),
					}); e != nil {
						a.Log.Error(`fail Modify WamStatusDaily schedule: ` + e.Error())
					}
				}
			})

		a.cronDaily.StartAsync()
	}

}
func (a *Acc) StopDailyCron() {
	if a.cronDaily != nil {
		a.cronDaily.Stop()
		a.cronDaily.Clear()
		a.cronDaily = nil
		a.Wg.Done()
	}
}

// send w:stats
func (a *Acc) StartWamCron() {
	if a.cronWam == nil {
		a.Wg.Add(1)
		a.cronWam = gocron.NewScheduler(time.UTC)

		a.cronWam.Every(1).Minutes().Do(func() {
			if arand.Int(0, 10) != 0 { // 1/10 percent
				return
			}
			e := a.wam_stats()
			if e != nil {
				a.Log.Error("fail wam cron: " + e.Error())
			}
		})
		a.cronWam.StartAsync()
	}
}
func (a *Acc) StopWamCron() {
	if a.cronWam != nil {
		a.cronWam.Stop()
		a.cronWam.Clear()
		a.cronWam = nil
		a.Wg.Done()
	}

}

// send media_conn
func (a *Acc) StartMediaConnCron() {
	if a.cronMediaConn == nil {
		a.Wg.Add(1)
		a.cronMediaConn = gocron.NewScheduler(time.UTC)

		// start 1 hour later, runs every 1 hour
		a.cronMediaConn.Every(1).Hour().StartAt(time.Now().Add(time.Hour)).Do(func() {
			e := a.set_media_conn()
			if e != nil {
				a.Log.Error("fail media_conn cron: " + e.Error())
			}
		})
		a.cronMediaConn.StartAsync()
	}
}
func (a *Acc) StopMediaConnCron() {
	if a.cronMediaConn != nil {
		a.cronMediaConn.Stop()
		a.cronMediaConn.Clear()
		a.cronMediaConn = nil
		a.Wg.Done()
	}

}

// stop all cron
func New_Hook_StopCron(a *Acc) func(...any) error {
	return func(...any) error {
		a.StopPingCron()
		a.StopWamCron()
		a.StopDailyCron()
		a.StopMediaConnCron()
		return nil
	}
}
