package net

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"ahex"
	"algo"
	"event"
	"phoenix"
	"scope"
	"wa/db"
	"wa/def"
	"wa/noise"
	"wa/pb"
	"wa/xmpp"

	"github.com/fanliao/go-promise"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type NoiseSocket struct {
	Socket
	*event.Event[string]

	pingSchedule *gocron.Scheduler

	MtxConnected sync.Mutex
	mtxCs        sync.RWMutex
	cs_me        *noise.CipherState
	cs_svr       *noise.CipherState

	mtxPool sync.RWMutex
	pool    map[string]*promise.Promise

	wg sync.WaitGroup

	mtxIqId_1 sync.Mutex
	IqId_1    uint16 // begin from "1"

	mtxIqId_2 sync.Mutex
	IqId_2    uint16 // begin from "00"
}

func NewNoiseSocket(
	ev *event.Event[string],
	proxy string,
	dns map[string]string,
) *NoiseSocket {
	ns := &NoiseSocket{
		Socket: Socket{Proxy: proxy, Dns: dns},
		Event:  ev,
		pool:   make(map[string]*promise.Promise),
	}

	ns.ResetIqId()

	ev.On(def.Ev_Connected, func(...any) error {
		go ns.StartReading()

		return nil
	})

	ev.On(def.Ev_NoiseDisconnected, func(...any) error {
		ns.cancelAllPromise() // cancel all read waiting
		// clean up
		ns.mtxCs.Lock()
		ns.cs_me = nil
		ns.cs_svr = nil
		ns.mtxCs.Unlock()

		return nil
	})

	return ns
}

func (this *NoiseSocket) IsConnected() bool {
	this.mtxCs.RLock()
	defer this.mtxCs.RUnlock()

	return this.cs_me != nil
}
func (this *NoiseSocket) NextIqId_1() string {
	this.mtxIqId_1.Lock()
	defer this.mtxIqId_1.Unlock()

	curr := this.IqId_1
	this.IqId_1++
	return fmt.Sprintf("%x", curr)
}
func (this *NoiseSocket) NextIqId_2() string {
	this.mtxIqId_2.Lock()
	defer this.mtxIqId_2.Unlock()

	curr := this.IqId_2
	this.IqId_2++

	return "0" + fmt.Sprintf("%x", curr)
}

func (this *NoiseSocket) WriteRoutingInfo(routingInfo []byte) error {
	this.Socket.EnableWriteTimeout(true)
	defer this.Socket.EnableWriteTimeout(false)

	e := this.Socket.write_n(def.ED_01)
	if e != nil {
		return e
	}
	this.Event.Fire(def.Ev_Log, db.DEBUG, "sending routing info: \n"+hex.Dump(routingInfo))
	return this.Socket.WritePacket(routingInfo)
}
func (this *NoiseSocket) ResetIqId() {
	this.IqId_1 = 1 // begin from "1", "0" is noise success
	this.IqId_2 = 0 // begin from "00"
}

// XX:
// -> e
// <- e, ee, s, es
// -> s, se
func (this *NoiseSocket) HandshakeXX(
	staticI noise.DHKey,
	devProto []byte,
	routingInfo []byte,
) ([]byte, error) {

	this.ResetIqId()

	if e := this.Close(); e != nil {
		return nil, errors.Wrap(e, `fail Close`)
	}
	if e := this.Socket.Connect(); e != nil {
		return nil, errors.Wrap(e, `Socket.Connect`)
	}
	this.Socket.EnableReadTimeout(true)
	defer this.Socket.EnableReadTimeout(false)
	this.Socket.EnableWriteTimeout(true)
	defer this.Socket.EnableWriteTimeout(false)

	// 0. prepare
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)

	hs, e := noise.NewHandshakeState(noise.Config{
		StaticKeypair: staticI,
		CipherSuite:   cs,
		Pattern:       noise.HandshakeXX,
		Initiator:     true,
		Prologue:      def.WA_41,
	})
	if e != nil {
		return nil, errors.New(`handshake st`)
	}

	// 0. Prologue
	// ED_01
	if len(routingInfo) > 0 {
		if e := this.WriteRoutingInfo(routingInfo); e != nil {
			return nil, errors.Wrap(e, "write routing_info")
		}
	}
	// WA_41
	if e := this.Socket.write_n(def.WA_41); e != nil {
		return nil, errors.Wrap(e, `write wa41`)
	}

	// 1 write: e ->
	_, _, _, _ = hs.WriteMessage(nil, nil)
	bs, _ := proto.Marshal(&pb.PatternXX{
		E: &pb.PatternStep{
			S1: hs.LocalEphemeral().Public[:],
		},
	})

	if e := this.Socket.WritePacket(bs); e != nil {
		return nil, errors.Wrap(e, `write msg 1`)
	}

	// 2.1 read e, es, s, ss <-
	pkt, e := this.Socket.ReadPacket()
	if e != nil {
		return nil, errors.Wrap(e, `sock read msg 1`)
	}
	xx2 := pb.PatternXX{}
	e = proto.Unmarshal(pkt, &xx2)
	if e != nil {
		return nil, errors.Wrap(e, `unmarshal msg 1`)
	}
	p123 := []byte{}
	p123 = append(p123, xx2.E_EE_S_ES.S1...)
	p123 = append(p123, xx2.E_EE_S_ES.S2...)
	p123 = append(p123, xx2.E_EE_S_ES.S3...)
	_, _, _, e = hs.ReadMessage(nil, p123)
	if e != nil {
		return nil, errors.Wrap(e, `hs read msg 1`)
	}

	// 3 write es, ss ->
	this.mtxCs.Lock()
	pkt, this.cs_me, this.cs_svr, e = hs.WriteMessage(nil, devProto)
	this.mtxCs.Unlock()

	sg := scope.Guard{Fn: func() {
		this.cs_me = nil
		this.cs_svr = nil
	}}
	defer sg.Exec()

	if e != nil {
		return nil, errors.Wrap(e, `write msg 2`)
	}
	xx3 := &pb.PatternXX{
		S_SE: &pb.PatternStep{S1: pkt[:0x30], S2: pkt[0x30:]},
	}
	bs, _ = proto.Marshal(xx3)
	if e = this.Socket.WritePacket(bs); e != nil {
		return nil, errors.Wrap(e, `WritePacket 2`)
	}

	// 4 read 'success'
	node, e := this.readXmppNode()
	if e != nil {
		return nil, errors.Wrap(e, `read final node`)
	}
	//fmt.Println("RemoteStatic")
	//fmt.Println(hex.Dump(hs.PeerStatic()))
	if node.Tag != `success` {
		return nil, errors.New(node.ToString())
	}
	if loc, ok := node.GetAttr(`location`); ok {
		this.Fire(def.Ev_Noise_Location, loc)
	}
	sg.Dismiss()

	return hs.PeerStatic(), nil
}

// IK:
// <- s
// ...
// -> e, es, s, ss
// <- e, ee, se
func (this *NoiseSocket) HandshakeIK(
	staticI noise.DHKey, staticR []byte,
	devProto []byte,
	routingInfo []byte,
) error {

	this.ResetIqId()

	if e := this.Close(); e != nil {
		return errors.Wrap(e, `fail Close`)
	}
	if e := this.Socket.Connect(); e != nil {
		return errors.Wrap(e, `Socket.Connect`)
	}
	this.Socket.EnableReadTimeout(true)
	defer this.Socket.EnableReadTimeout(false)
	this.Socket.EnableWriteTimeout(true)
	defer this.Socket.EnableWriteTimeout(false)

	// 0. prepare
	cs := noise.NewCipherSuite(noise.DH25519, noise.CipherAESGCM, noise.HashSHA256)

	hs, e := noise.NewHandshakeState(noise.Config{
		StaticKeypair: staticI,
		CipherSuite:   cs,
		Pattern:       noise.HandshakeIK,
		Initiator:     true,
		Prologue:      def.WA_41,
		PeerStatic:    staticR,
	})
	if e != nil {
		return errors.New(`handshake st`)
	}

	// 0. Prologue
	// ED_01
	if e := this.WriteRoutingInfo(routingInfo); e != nil {
		return errors.Wrap(e, "write routing_info")
	}
	// WA_41
	if e := this.Socket.write_n(def.WA_41); e != nil {
		return e
	}

	// 1. write
	//    e, es, s, ss ->
	pkt, _, _, e := hs.WriteMessage(nil, devProto)
	if e != nil {
		return e
	}
	ik1 := &pb.PatternIK{
		E_EE_S_ES: &pb.PatternStep{
			S1: pkt[:0x20], S2: pkt[0x20:0x50], S3: pkt[0x50:]},
	}
	bs, _ := proto.Marshal(ik1)
	if e = this.Socket.WritePacket(bs); e != nil {
		return e
	}

	// 2. read
	//    <- e, ee, se
	pkt, e = this.Socket.ReadPacket()
	if e != nil {
		return e
	}
	ik2 := pb.PatternIK{}
	if e = proto.Unmarshal(pkt, &ik2); e != nil {
		return e
	}
	p12 := []byte{}
	p12 = append(p12, ik2.E_EE_SE.S1...)
	p12 = append(p12, ik2.E_EE_SE.S3...)
	this.mtxCs.Lock()
	_, this.cs_me, this.cs_svr, e = hs.ReadMessage(nil, p12)
	this.mtxCs.Unlock()

	sg := scope.Guard{Fn: func() {
		this.cs_me = nil
		this.cs_svr = nil
	}}
	defer sg.Exec()

	if e != nil {
		return e
	}

	// 3. read 'success'
	node, e := this.readXmppNode()
	if e != nil {
		return e
	}
	if node.Tag != `success` {
		return errors.New(node.ToString())
	}
	if loc, ok := node.GetAttr(`location`); ok {
		this.Fire(def.Ev_Noise_Location, loc)
	}
	sg.Dismiss()

	return nil
}
func (this *NoiseSocket) encrypt(plain []byte) ([]byte, error) {
	this.mtxCs.Lock()
	defer this.mtxCs.Unlock()
	if this.cs_me == nil {
		return nil, errors.New(`not connected cs_me, maybe Connect() failed?`)
	}

	return this.cs_me.Encrypt(nil, nil, plain), nil
}
func (this *NoiseSocket) decrypt(cipher []byte) ([]byte, error) {
	this.mtxCs.Lock()
	defer this.mtxCs.Unlock()

	if this.cs_svr == nil {
		return nil, errors.New(`not connected cs_svr`)
	}
	return this.cs_svr.Decrypt(nil, nil, cipher)
}

func (this *NoiseSocket) decryptXmppNode(pkt []byte) (*xmpp.Node, error) {
	bs, e := this.decrypt(pkt)
	if e != nil {
		return nil, errors.Wrap(e, `fail decrypt`)
	}
	return this.parseXmppNode(bs)
}
func (this *NoiseSocket) parseXmppNode(pkt []byte) (*xmpp.Node, error) {
	var e error
	if len(pkt) < 4 {
		return nil, errors.New(`too short pkt`)
	}
	if pkt[0] == 2 {
		pkt, e = algo.UnZlib(pkt[1:])
		if e != nil {
			return nil, errors.New(`fail to unzlib`)
		}
	}
	if pkt[0] == 0 {
		pkt = pkt[1:]
	}
	r := xmpp.NewReader(pkt)
	n, e := r.ReadNode()
	if e != nil {
		return nil, errors.Wrap(e, `ReadNode fail: `+ahex.Enc(pkt))
	}
	this.Event.Fire(def.Ev_Log, db.DEBUG, "Read Node:\n%s", n.ToString())
	return n, nil
}

// write encrypted pkt
func (this *NoiseSocket) writePacket(plain []byte) error {
	cipher, e := this.encrypt(plain)
	if e != nil {
		return e
	}

	return this.Socket.WritePacket(cipher)
}
func (this *NoiseSocket) WriteXmppNode(n *xmpp.Node) error {
	bf := xmpp.NewWriter().WriteNode(n)

	if n.Compressed {
		bf = algo.Zlib(bf)
		bf = append([]byte{2}, bf...)
	} else {
		bf = append([]byte{0}, bf...)
	}

	this.Event.Fire(def.Ev_Log, db.DEBUG, "Send node:\n%s", n.ToString())
	e := this.writePacket(bf)
	if e != nil {
		this.Event.Fire(def.Ev_Log, db.ERROR, "err: %s", e.Error())
	}
	return e
}

func get_id_from_node(n *xmpp.Node) string {
	for _, attr := range n.Attrs {
		if attr.Key == `id` {
			return attr.Value
		}
	}
	return ``
}

func (this *NoiseSocket) cancelAllPromise() {
	this.mtxPool.Lock()
	defer this.mtxPool.Unlock()

	for _, p := range this.pool {
		p.Cancel()
	}
}

func (this *NoiseSocket) waitForIqId(iq_id string) *promise.Promise {
	p := promise.NewPromise()

	this.mtxPool.Lock()
	defer this.mtxPool.Unlock()

	this.pool[strings.ToLower(iq_id)] = p
	return p
}

func (this *NoiseSocket) dismissIqId(iq_id string) {
	this.mtxPool.Lock()
	defer this.mtxPool.Unlock()

	delete(this.pool, strings.ToLower(iq_id))
}
func (this *NoiseSocket) isWaitingForIqId(iq_id string) (*promise.Promise, bool) {
	this.mtxPool.RLock()
	defer this.mtxPool.RUnlock()

	p, ok := this.pool[strings.ToLower(iq_id)]
	return p, ok
}
func (this *NoiseSocket) isWaitingForPacket() (*promise.Promise, bool) {
	return this.isWaitingForIqId("")
}
func (this *NoiseSocket) readXmppNode() (*xmpp.Node, error) {
	bs, e := this.Socket.ReadPacket()
	if e != nil {
		return nil, e
	}

	n, e := this.decryptXmppNode(bs)
	if e != nil {
		return nil, e
	}
	return n, nil
}

func (this *NoiseSocket) waitXmppNode(id string) (*xmpp.Node, error) {
	p := this.waitForIqId(id)
	defer this.dismissIqId(id)

	resp, e, timeout := p.GetOrTimeout(uint(def.NET_TIMEOUT * 1000)) // millisec
	if e != nil {
		return nil, e
	}
	if timeout {
		return nil, errors.New(`get response timeout`)
	}
	return resp.(*xmpp.Node), nil
}

// send then receive Node with same `id`
func (this *NoiseSocket) WriteReadXmppNode(n *xmpp.Node) (*xmpp.Node, error) {
	iq_id := get_id_from_node(n)

	if e := this.WriteXmppNode(n); e != nil {
		return nil, e
	}
	return this.waitXmppNode(iq_id)
}

func (this *NoiseSocket) Close() error {
	this.Event.Fire(def.Ev_Log, db.DEBUG, "NoiseSocket.Socket.Close()")
	e := this.Socket.Close()
	// wait for write/read done
	this.Event.Fire(def.Ev_Log, db.DEBUG, "NoiseSocket.Socket.Close() return")

	this.Event.Fire(def.Ev_Log, db.DEBUG, "NoiseSocket.Close() now wg.Wait()")
	this.wg.Wait()
	this.Event.Fire(def.Ev_Log, db.DEBUG, "NoiseSocket.Close() wg.Wait() done")

	return e
}

func (this *NoiseSocket) StartReading() {
	defer phoenix.Ignore(nil)

	this.wg.Add(1)
	defer this.wg.Done()
	defer this.Event.Fire(def.Ev_Log, db.DEBUG, "NoiseSocket.StartReading returned")

	for {
		bs, e := this.Socket.ReadPacket()

		if e != nil {
			this.Event.Fire(def.Ev_Log, db.WARNING,
				"err disconnected, read fail from wa/proxy: %s", e.Error())
			this.Event.Fire(def.Ev_NoiseDisconnected, errors.Wrap(e, "WA_Disconnected"))
			return
		}

		if !this.IsConnected() { // not handshake yet
			p, is_waiting := this.isWaitingForPacket()
			if is_waiting { // waiting for the node
				p.Resolve(bs)
				continue
			}
		} else { // done handshake
			n, e := this.decryptXmppNode(bs)
			if e != nil {
				this.Event.Fire(def.Ev_Log, db.ERROR, `fail decrypt: `+e.Error())
				continue
			}

			iq_id := get_id_from_node(n)
			iq_id = strings.ToLower(iq_id)
			p, is_waiting := this.isWaitingForIqId(iq_id)
			if is_waiting { // waiting for the node
				p.Resolve(n)
				continue
			}

			/*
			 hooks may stop event propagation
			 eg: ping response is ignored to client
			 Note: n maybe modified
			*/
			ev := this.Event.Fire(n.Tag, n) // trigger hooks
			if errors.Is(ev, event.Stop) {  // if killed by hook, not push to client
				//this.Log.Debugf("Node processed by hook: %s", n.ToString())
			} else {
				this.Event.Fire(`push`, n)
			}
		}
	}
}
