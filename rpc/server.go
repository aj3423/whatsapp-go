package rpc

import (
	"context"
	"reflect"

	"ajson"
	"phoenix"
	"wa/core"
	"wa/def"
	"wa/rpc/pb"
	"wa/xmpp"

	"github.com/pkg/errors"
)

type WaServer struct {
	pb.UnimplementedRpcServer
}

// required field: `acc`, `func`
func (s *WaServer) Exec(ctx context.Context, in *pb.Json) (ret *pb.Json, e error) {
	defer phoenix.Ignore(func() {
		ret = &pb.Json{Data: core.NewCrashRet().ToString()}
	})

	j, e := ajson.Parse(in.GetData())
	if e != nil {
		e = errors.Wrap(e, `Fail parse input json`)
		return
	}

	func_ := j.Get(`func`).String()

	// check if core.xxx exists
	fn := reflect.ValueOf(&core.CORE).MethodByName(func_)
	if !fn.IsValid() {
		e = errors.New(`invalid func: ` + func_)
		return
	}

	// call core.xxx functions
	args := []reflect.Value{reflect.ValueOf(j)}
	r := fn.Call(args)
	rj := r[0].Interface().(*ajson.Json)

	ret = &pb.Json{Data: rj.ToString()}
	return
}

func (s *WaServer) Push(in *pb.Json, stream pb.Rpc_PushServer) (e error) {
	defer phoenix.Ignore(func() {
		e = errors.New("Push crashed")
	})

	j, e := ajson.Parse(in.GetData())
	if e != nil {
		e = errors.Wrap(e, `Fail parse input json`)
		return
	}
	// get acc
	a, e := core.GetAccFromJson(j)
	if e != nil {
		return
	}

	a.Log.Debug("Push.1 try a.Lock")
	a.Lock()
	a.Log.Debug("Push.2 a.Lock ok")

	// prevent connect multiple push rpc
	if a.PushStreamConnected {
		e = errors.New("acc.pushStream already exists")
		a.UnLock() // important !!!!!!!!!
		return
	}

	a.Log.Debug("new Push stream connected")
	defer a.Log.Debug("Push stream end")

	a.Wg.Add(1)
	defer a.Wg.Done()

	// push notification
	push_to_client := func(args ...any) error {
		j := args[0].(*xmpp.Node).ToJson()
		e := stream.Send(&pb.Json{
			Data: core.NewJsonRet(j).ToString(),
		})
		if e != nil {
			a.Log.Error("stream push fail: " + e.Error())
		}
		return nil
	}
	sub := a.Event.On(def.Ev_Push, push_to_client)
	defer a.Event.Un(sub)

	// the long link disconnects when this function returns
	// so don't return, blocking to keep alive
	// until client disconnected or AccOff()
	stm := NewStream(stream)

	// disconnected
	disconnected := func(args ...any) error {
		a.Log.Debug("push 'disconnect' to client")
		// send last error msg
		stream.Send(&pb.Json{
			Data: core.NewErrRet(args[0].(error)).ToString(),
		})
		return nil
	}
	sub_1 := a.Event.On(def.Ev_NoiseDisconnected, disconnected)
	defer a.Event.Un(sub_1)

	// close only when  AccOff
	close := func(...any) error {
		a.Log.Debug("on acc_off, close stream")
		stm.Close()
		return nil
	}
	sub_2 := a.Event.On(def.Ev_AccOff, close)
	defer a.Event.Un(sub_2)

	a.PushStreamConnected = true
	defer func() { a.PushStreamConnected = false }()

	// done event binding, release the event lock
	a.UnLock()

	a.Log.Debug("stream KeepAlive()")
	stm.KeepAlive() // blocking, wait for stm.Close()

	a.Log.Info("push closed")

	return
}
