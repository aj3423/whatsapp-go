package core

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"afs"
	"ajson"
	"algo"
	"event"
	"run"

	"go.mongodb.org/mongo-driver/bson"

	"wa/db"
	"wa/def"
	"wa/net"
	"wa/wam"

	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
)

type Acc struct {
	Noise *net.NoiseSocket
	Store *db.Store

	// for awei bug: AccOff/Push stream at same time
	// causes AccOff hangs forever
	// anything changes event binding to acc, can use this lock
	mu sync.Mutex

	*event.Event[string]

	Wg sync.WaitGroup // used in AccOff, ==zero then delete Acc object

	PushStreamConnected bool // gprc client stream

	cronPing      *gocron.Scheduler
	cronWam       *gocron.Scheduler
	cronDaily     *gocron.Scheduler
	cronMediaConn *gocron.Scheduler

	Log *db.Logger
}

func (a *Acc) Lock() {
	a.mu.Lock()
}
func (a *Acc) UnLock() {
	a.mu.Unlock()
}

var mu_acc = sync.RWMutex{}
var mapAcc = map[uint64]*Acc{}

func NewAcc(acc_id uint64) (*Acc, error) {
	ev := event.New[string]()

	// store
	store := db.NewStore(acc_id)

	// proxy
	proxy, dns, e := store.GetProxy() // it's ok if acc not exists in db
	if e != nil {
		return nil, e
	}

	a := &Acc{
		Store: store,
		Log:   &db.Logger{AccId: acc_id},
		Event: ev,
		Noise: net.NewNoiseSocket(ev, proxy, dns),
	}

	ev.On(def.Ev_Log, NewLogHook(a))

	ev.On(def.Ev_AccCreate, New_Hook_AccCreate(a))

	// receive call
	ev.On(def.Ev_call, New_Hook_CallOffer(a))
	ev.On(def.Ev_call, New_Hook_CallAck(a))
	// save location on last resp of Connect
	ev.On(def.Ev_Noise_Location, New_Hook_NoiseLocation(a))
	// send ping every 4 min
	ev.On(def.Ev_Heartbeat, New_Hook_Ping(a))
	// stop all cron job when disconnected
	ev.On(def.Ev_NoiseDisconnected, New_Hook_StopCron(a))
	// urn:xmpp:ping
	ev.On(def.Ev_iq, New_Hook_ServerPing(a))
	// decode message
	ev.On(def.Ev_message, New_Hook_Msg(a))
	ev.On(def.Ev_message, New_Hook_GroupMsg(a))
	// receipt ack
	ev.On(def.Ev_receipt, New_Hook_Receipt(a))
	// dirty (group/account_sync)
	ev.On(def.Ev_ib, New_Hook_Dirty(a)) // cause server stop pushing when XX
	// safetynet (attestation/verify_apps)
	ev.On(def.Ev_ib, New_Hook_Safetynet(a))
	// save routing_info for connecting
	ev.On(def.Ev_ib, New_Hook_RoutingInfo(a))

	// multi device(web login)
	ev.On(def.Ev_notification, New_Hook_MultiDeviceAdd(a))
	ev.On(def.Ev_notification, New_Hook_MultiDeviceRemove(a))
	ev.On(def.Ev_notification, New_Hook_MultiDeviceUpdate(a))

	// group
	ev.On(def.Ev_notification, New_Hook_GroupCreate(a))
	ev.On(def.Ev_notification, New_Hook_GroupAdd(a))
	ev.On(def.Ev_notification, New_Hook_GroupLeave(a))
	// Peer SetEncrypt,clear session/identity
	ev.On(def.Ev_notification, New_Hook_PeerIdentityChange(a))
	// server ask for more prekeys
	ev.On(def.Ev_notification, New_Hook_NeedMorePrekeys(a))
	// send ack to all notifications
	ev.On(def.Ev_notification, New_Hook_Notification(a))

	exist_in_db, e := store.AccExists()
	if e != nil {
		return nil, e
	}
	if !exist_in_db {
		e = ev.Fire(def.Ev_AccCreate)
		if e != nil {
			return nil, e
		}
	}

	return a, nil
}

func (c Core) AccOn(j *ajson.Json) *ajson.Json {
	acc_id, e := GetAccIdFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	// if exists in memory, return
	{
		_, ok := get_memory_acc(acc_id)

		if ok {
			return NewErrRet(errors.New(`acc already on`))
		}
	}

	// not exists in memory, new acc instance
	a, e := NewAcc(acc_id)
	if e != nil {
		return NewErrRet(e)
	}
	a.Log.Debug("AccOn")

	// store in memory
	if e := set_memory_acc(acc_id, a); e != nil {
		return NewErrRet(e)
	} else {
		return NewSucc()
	}
}
func (c Core) AccInfo(j *ajson.Json) *ajson.Json {
	acc_id, e := GetAccIdFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	_, isOn := get_memory_acc(acc_id)

	rj := NewSucc()

	rj.Set(`isOn`, isOn)
	if isOn { // if isOn, then must exists
		rj.Set(`exists`, true)
	} else {
		sto := db.NewStore(acc_id)
		exists, e := sto.AccExists()
		if e != nil {
			return NewErrRet(e)
		}
		rj.Set(`exists`, exists)
	}

	return rj
}
func (c Core) AccDel(j *ajson.Json) *ajson.Json {
	id, e := GetAccIdFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	_, ok := get_memory_acc(id)
	if ok {
		return NewErrRet(errors.New(`acc is on, AccOff first`))
	}

	e = db.DeleteAcc(id)
	if e != nil {
		return NewErrRet(errors.New(`fail delete acc`))
	} else {
		return NewSucc()
	}
}

func (c Core) AccOff(j *ajson.Json) *ajson.Json {
	id, e := GetAccIdFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	a, ok := get_memory_acc(id)
	if !ok {
		return NewSucc()
	}
	a.Log.Debug("AccOff 1, try acc.Lock()")
	a.Lock()
	defer a.UnLock()

	a.Log.Debug("AccOff 2, acc locked")
	defer a.Log.Debug("AccOff done")

	// clean up
	{
		a.Log.Debug("AccOff.Noise.Close()")
		// must close first, close uses events
		a.Noise.Close()
		a.Log.Debug("AccOff.Fire.(Ev_AccOff)")
		a.Event.Fire(def.Ev_AccOff) // close rpc stream
		a.Log.Debug("AccOff.Event.Clear()")
		a.Event.Clear()

		a.Log.Debug("AccOff.Wait")
		a.Wg.Wait() // wait for rpc push finish
	}

	a.Log.Debug("delete_memory_acc")

	// do delete
	delete_memory_acc(id)

	return NewSucc()
}

func get_memory_acc(accid uint64) (*Acc, bool) {
	mu_acc.RLock()
	s, ok := mapAcc[accid]
	mu_acc.RUnlock()
	return s, ok
}
func set_memory_acc(id uint64, acc *Acc) error {
	if _, ok := get_memory_acc(id); ok {
		return errors.New(`acc already on, call 'AccOff' first`)
	}

	mu_acc.Lock()
	mapAcc[id] = acc
	mu_acc.Unlock()

	return nil
}
func delete_memory_acc(accid uint64) {
	mu_acc.Lock()
	delete(mapAcc, accid)
	mu_acc.Unlock()
}
func GetAccIdFromJson(j *ajson.Json) (uint64, error) {
	if u64, e := j.Get(`acc`).TryUint64(); e == nil { // is u64
		return u64, nil
	}
	if acc_id_str, e := j.Get(`acc`).TryString(); e == nil { // is string
		if u64, e := strconv.ParseUint(acc_id_str, 10, 64); e == nil {
			return u64, nil
		}
	}
	return 0, errors.New(`fail parse 'acc' id`)
}
func GetAccFromJson(j *ajson.Json) (*Acc, error) {
	id, e := GetAccIdFromJson(j)
	if e != nil {
		return nil, e
	}
	a, ok := get_memory_acc(id)
	if ok {
		return a, nil
	} else {
		return nil, errors.New(`acc not on`)
	}
}

func (c Core) ServerStatus(j *ajson.Json) *ajson.Json {
	rj := NewSucc()

	ids := []string{}

	mu_acc.RLock()
	for id := range mapAcc {
		ids = append(ids, strconv.Itoa(int(id)))
	}
	mu_acc.RUnlock()

	rj.Set("acc_on", ids)

	return rj
}

func (c Core) AccDump(j *ajson.Json) *ajson.Json {
	id, e := GetAccIdFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	dir := fmt.Sprintf("acc_dump_%d", id)

	tables := []string{"Profile", "Device", "Config", "Schedule", "Session", "Prekey", "Identity", "SignedPrekey", "SenderKey", "Proxy", "Message", "Group", "GroupMember", "WamSchedule", "Cdn"}

	defer func() {
		afs.RemoveDir(dir)
	}()
	for _, table := range tables {
		stdout, stderr, ec := run.RunCommand(".", "mongodump",
			"--db", "wa",
			"--query", `{"AccId":`+fmt.Sprintf("%d", id)+`}`,
			"--out", dir,
			"--collection", table,
		)
		if ec != 0 {
			return NewErrRet(errors.New(fmt.Sprintf("fail run mongodump, stdout: %s, stderr: %s", stdout, stderr)))
		}
	}
	zip := dir + ".zip"
	stdout, stderr, ec := run.RunCommand(".", "zip", zip, dir)
	if ec != 0 {
		return NewErrRet(errors.New(fmt.Sprintf("fail run zip, stdout: %s, stderr: %s", stdout, stderr)))
	}
	defer func() {
		afs.Remove(zip)
	}()

	bs := afs.Read(zip)
	rj := NewSucc()
	rj.Set("zip", algo.B64Enc(bs))
	return rj
}

func New_Hook_AccCreate(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(...any) error {

		e := a.Store.FirstInit()
		if e != nil {
			return e
		}

		// wam EulaAccept
		{
			e := a.Store.ModifyWamEvent(bson.M{
				`EulaAccept`: time.Now(),
			})
			if e != nil {
				return e
			}
		}

		// wam PsIdCreate
		{
			e := a.AddWamEventBuf(wam.WamPsIdCreate, func(cc *wam.ClassChunk) {
			})
			if e != nil {
				return e
			}
		}
		return nil
	}
}
