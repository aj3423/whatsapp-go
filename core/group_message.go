package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"ahex"
	"ajson"
	"algo"
	"arand"
	"event"
	"phoenix"
	"wa/crypto"
	"wa/def"
	"wa/pb"
	"wa/signal/groups"
	"wa/signal/protocol"
	"wa/signal/session"
	"wa/xmpp"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

func build_hash_jid(jid string) string {
	recid, devid, _ := split_jid(jid)
	return fmt.Sprintf("%d.0:%d@s.whatsapp.net", recid, devid)
}

// 111.0:2@s.whatsapp.net -> [
//
//	111.0:0@s.whatsapp.net,
//	111.0:1@s.whatsapp.net,
//	111.0:2@s.whatsapp.net
//
// ]
func get_all_device_jid(jid string) (ret []string) {
	p := strings.Index(jid, ":")
	if p < 0 {
		ret = []string{jid}
	} else {
		n, _ := strconv.Atoi(jid[p+1 : p+2])
		for i := 0; i <= n; i++ {
			ret = append(ret, jid[0:p+1]+strconv.Itoa(i)+jid[p+2:])
		}
	}
	return
}

func phash(me string, participants []string) string {
	s := []string{}

	for _, jid := range get_all_device_jid(me) {
		s = append(s, build_hash_jid(jid))
	}

	for _, part := range participants {
		s = append(s, build_hash_jid(part))
	}

	sort.Strings(s)

	str := strings.Join(s, "")

	sha := algo.Sha256([]byte(str))
	return `2:` + algo.B64Enc(sha[:6])
}

func find_whisper_node(n *xmpp.Node) (*xmpp.Node, uint32, bool) {
	var msg_type string
	wmNode, has_wm := n.FindChild(func(ch *xmpp.Node) bool {
		msg_type, _ = ch.GetAttr(`type`)
		return msg_type == `pkmsg` || msg_type == `msg`
	})
	if has_wm {
		return wmNode, msg_type_int(msg_type), true
	} else {
		return nil, 0, false
	}
}
func find_sk_node(n *xmpp.Node) (*xmpp.Node, bool) {
	return n.FindChildWithAttr(`type`, `skmsg`)
}

func (c Core) SendGroupMsg(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	media_t := MediaTypeInt(j.Get(`media_type`).String())
	if media_t == pb.Media_Unknown {
		return NewErrRet(errors.New(`unsupported media_type`))
	}
	// parse media from json
	media, e := NewMedia(media_t)
	if e != nil {
		return NewErrRet(e)
	}
	mj, ok := j.TryGet(`media`)
	if !ok {
		return NewErrRet(errors.New(`fail parse media json`))
	}
	// parse Media from json
	if media.FillFromJson(mj) != nil {
		return NewErrRet(errors.New(`fail parse media json`))
	}

	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}

	gid := j.Get(`gid`).String()

	participants, e := j.Get(`participants`).TryStringArray()
	if e != nil {
		participants, e = a.Store.ListGroupMemberJid(gid, false)
		if e != nil {
			return NewErrRet(e)
		}
	}

	// expand all jid with multi-device
	participants, e = expand_jids_devices(a, participants)
	if e != nil {
		return NewErrRet(e)
	}

	recid_me := dev.Cc + dev.Phone
	addr_me := protocol.NewSignalAddress(recid_me, 0)

	skn_me := protocol.NewSenderKeyName(gid, addr_me)

	map_sb, e := a.ensure_session_builder(participants)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail ensure_session_builder`))
	}

	gsb := groups.NewGroupSessionBuilder(a.Store)

	grp_cipher := groups.NewGroupCipher(gsb, skn_me, a.Store)

	// sender key distribution msg
	skdm, e := gsb.Create(skn_me)
	if e != nil {
		return NewErrRet(e)
	}

	// sender key msg
	pbm := &pb.Message{}
	media.FillMessage(pbm)

	p, _ := proto.Marshal(pbm)
	padded := crypto.RandomPadMsg(p)
	skm_, e := grp_cipher.Encrypt(padded)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail grp_cipher.Encrypt`))
	}
	skm := skm_.(*protocol.SenderKeyMessage)

	// participants
	ptcps := &xmpp.Node{
		Tag: `participants`,
	}
	for _, jid := range participants {
		sb := map_sb[jid] // guaranteed ok

		// build protobuf + pad
		pbm := pb.Message{
			Grp: &pb.Message_Group{
				Id:   &gid,
				Skdm: skdm.Serialize(),
			},
		}
		p, _ := proto.Marshal(&pbm)
		padded := crypto.RandomPadMsg(p)

		// session cipher
		recid, devid, e := split_jid(jid)
		if e != nil {
			return NewErrRet(e)
		}
		peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)
		sc := session.NewCipher(sb, peer_addr)

		// encrypt msg
		encMsg, e := sc.Encrypt(padded)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail encrypt msg`))
		}

		child := &xmpp.Node{
			Tag: `to`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
			Children: []*xmpp.Node{
				{
					Tag: `enc`,
					Attrs: []*xmpp.KeyValue{
						{Key: `type`, Value: msg_type_str(encMsg.Type())},
						{Key: `v`, Value: `2`},
					},
					Data: encMsg.Serialize(),
				},
			},
		}
		ptcps.Children = append(ptcps.Children, child)
	}

	msg_id := strings.ToUpper(algo.Md5Str([]byte(arand.Uuid4())))

	cipher_type := `skmsg`
	cipher_ver := `2`
	attrs := []*xmpp.KeyValue{
		{Key: `type`, Value: cipher_type},
		{Key: `v`, Value: cipher_ver},
	}
	if media.MsgCategory() == `media` {
		attrs = append(attrs, &xmpp.KeyValue{
			Key: `mediatype`, Value: MediaTypeStr(media.Type()),
		})
	}

	// WamE2eMessageSend
	is_sns := gid == `status@broadcast`
	{
		var dest int32 = 1 // group
		if is_sns {
			dest = 3
		}
		e := a.wam_e2e_message_send(
			media_t,
			wam_cipher_text_type(cipher_type),
			wam_cipher_text_ver(cipher_ver),
			dest)
		if e != nil {
			a.Log.Error(`fail wam_e2e_message_send: ` + e.Error())
		}
	}

	send_begin := time.Now()

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `message`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: msg_id},
			{Key: `phash`, Value: phash(recid_me, participants)},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: media.MsgCategory()},
		},
		Children: []*xmpp.Node{
			{
				Tag:   `enc`,
				Attrs: attrs,
				Data:  skm.SignedSerialize(),
			},
			ptcps,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	// only save on success, cause I don't know the fail value for messageSendResult
	{
		international := false
		msg_type := 2 // 2: group
		if is_sns {
			msg_type = 4
		} else if strings.HasSuffix(gid, `@broadcast`) {
			msg_type = 3
		}
		er := a.wam_message_send(media_t, send_begin, msg_type, international)
		if er != nil {
			a.Log.Error(`fail WamMessageSend: ` + er.Error())
		}
	}
	return NewJsonRet(nr.ToJson())
}
func New_Hook_GroupMsg(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		// MUST have attr `id`,`participant`,`from`
		msg_id, ok1 := attrs[`id`]
		participant, ok2 := attrs[`participant`]
		gid, ok3 := attrs[`from`]
		t, ok4 := attrs[`t`]

		if !ok1 || !ok2 || !ok3 || !ok4 {
			return nil
		}

		// TODO, handle broadcast
		if strings.Contains(gid, `@broadcast`) {
			a.Store.EnsureMessage(
				msg_id, xmpp.NewWriter().WriteNode(n))
			a.receipt_group_msg_receive(msg_id, gid, participant)
			return nil
		}

		recid_peer, devid, e := split_jid(participant)
		if e != nil {
			return e
		}
		peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid_peer), devid)
		gsb := groups.NewGroupSessionBuilder(a.Store)
		skn := protocol.NewSenderKeyName(gid, peer_addr)

		// 1. get current wmNode/skNode
		wmNode, msg_type, has_wm := find_whisper_node(n)
		skNode, has_sk := find_sk_node(n)

		// if fail, send retry packet
		ev := event.New[string]()
		ev.On(def.Ev_Retry, func(...any) error {
			direct_distribution := !has_sk // only `pkmsg`
			if e := a.retry_group_msg(msg_id, attrs, direct_distribution); e != nil {

				// maybe caused by multi-dev, no time to handle that, just send a usync
				//if strings.Contains(participant, ":") { // multi device(maybe web client)
				//a.Wg.Add(1)
				//go func() {
				//a.usync_multi_device(participant)
				//a.Wg.Done()
				//}()
				//}

				// error means too much retry
				// just receipt or connection'll be lost
				return a.receipt_group_msg_receive(msg_id, gid, participant)
			} else {
				return event.Stop // if retry is working, kill event, don't push to client
			}
		})
		// if all success, send receipt packet
		ev.On(def.Ev_Success, func(args ...any) error {
			// send ack
			{
				a.receipt_group_msg_receive(msg_id, gid, participant)
			}

			content, _ := args[0].(*MessageContent)

			{ // log Message.proto
				if bs, e := proto.Marshal(content.P); e == nil {
					a.Log.Info("Message.Proto: " + ahex.Enc(bs))
				}
			}

			media := content.GetMedia()

			// wam
			{
				dev, e := a.Store.GetDev()
				if e != nil {
					return e
				}
				attrs := n.Children[0].MapAttrs()
				cipher_type, _ := attrs[`type`]
				cipher_ver, _ := attrs[`v`]

				timestamp, _ := strconv.Atoi(t)
				international := !strings.HasPrefix(participant, dev.Cc)

				if e := a.wam_e2e_message_recv(
					media.Type(),
					timestamp,
					wam_cipher_text_type(cipher_type),
					wam_cipher_text_ver(cipher_ver),
					1, /*dest*/
				); e != nil {
					a.Log.Error(`fail wam_e2e_message_recv: ` + e.Error())
				}

				if e := a.wam_message_receive(
					media.Type(),
					timestamp,
					2, /*msg_type*/
					international,
				); e != nil {
					a.Log.Error(`fail wam_message_receive: ` + e.Error())
				}
			}

			// modify node
			n.Attrs = append(n.Attrs, &xmpp.KeyValue{
				Key:   `media`,
				Value: media.ToJson().ToString(),
			})
			n.Attrs = append(n.Attrs, &xmpp.KeyValue{
				Key:   `media_type`,
				Value: MediaTypeStr(media.Type()),
			})

			// too long, not necessary to clients
			n.Children = n.Children[:0]

			return a.Store.SaveDecodedMessage(
				msg_id, uint32(media.Type()), media.Serialize())
		})

		// 0. save msg to db
		pm, e := a.Store.GetMessage(msg_id)
		if errors.Is(e, mongo.ErrNoDocuments) {
			pm, e = a.Store.EnsureMessage(
				msg_id, xmpp.NewWriter().WriteNode(n))
		}
		if e != nil {
			return e
		}
		// if already decrypted, return
		if pm.Decrypted {
			a.Log.Error("msg already decrypted: %s", msg_id)

			med, e := NewMediaFromBytes(
				pb.Media_Type(pm.MediaType), pm.DecMedia)
			if e != nil {
				return e
			}
			return ev.Fire(`success`, NewMessageContentFromDecrypted(
				med,
			))
		}

		// 2. process wmNode
		if has_wm {
			content, e := a.decodeGroupMsgSkdmFromJid(
				participant, wmNode.Data, msg_type)
			if e != nil {
				return ev.Fire(`retry`, nil)
			}
			// must have skdm
			skdm, e := content.Skdm()
			if e != nil {
				return ev.Fire(`retry`, nil)
			}
			if e := gsb.Process(skn, skdm); e != nil {
				return ev.Fire(`retry`, nil)
			}
			if content.GetMedia().Type() != pb.Media_Unknown {
				// if plain exists, then no need sk
				a.Log.Debug("wm contains plain,no need sk")
				return ev.Fire(`success`, content)
			}
		}
		if !has_sk {
			pn, e := xmpp.NewReader(pm.Node).ReadNode()
			if e != nil {
				a.Log.Error("(should never see this, report bug) fail read node from group msg: %s: %s", msg_id, e.Error())
				return ev.Fire(`retry`, nil)
			}
			skNode, has_sk = find_sk_node(pn)
			if !has_sk {
				a.Log.Error("(should never see this, report bug) group msg with no sk: %s", msg_id)
				return ev.Fire(`retry`, nil)
			}
		}

		// process sk node
		content, e := a.decodeSkMsg(gsb, skn, skNode.Data)
		if e != nil {
			return ev.Fire(`retry`, nil)
		}

		a.Log.Debug("sk contains plain")
		return ev.Fire(`success`, content)
	}
}

func (a *Acc) decodeGroupMsgSkdmFromJid(
	jid string,
	cipher []byte,
	msg_type uint32,
) (*MessageContent, error) {
	unpadded, e := a.decodeWspProtoFromJid(
		jid, cipher, msg_type)
	if e != nil {
		return nil, e
	}
	return parseMessageProto(unpadded)
}
func (a *Acc) decodeSkMsg(
	gsb *groups.SessionBuilder,
	skn *protocol.SenderKeyName,
	cipher []byte,
) (*MessageContent, error) {
	grp_cipher := groups.NewGroupCipher(gsb, skn, a.Store)

	skmsg, e := protocol.NewSenderKeyMessageFromBytes(cipher)
	if e != nil {
		return nil, e
	}

	bs, e := grp_cipher.Decrypt(skmsg)
	if e != nil { // retry on decode err
		return nil, e
	}
	unpadded, e := crypto.UnPadMsg(bs)
	if e != nil {
		return nil, e
	}

	return parseMessageProto(unpadded)
}

func (a *Acc) retry_group_msg(
	msg_id string,
	attrs map[string]string,
	direct_distribution bool,
) error {
	retry_times, e := a.Store.GetMessageRetry(msg_id)
	if e != nil {
		a.Log.Error("fail Get Retry times for msg %s: %s", msg_id, e.Error())
		return e
	}
	if retry_times >= MaxMessageRetry {
		eStr := fmt.Sprintf("too much retry for msg %s", msg_id)
		a.Log.Warning(eStr)
		return errors.New(eStr)
	}

	{ // retry
		a.Wg.Add(1)
		go func() {
			defer phoenix.Ignore(nil)
			a.retry_receipt(attrs, direct_distribution)
			a.Wg.Done()
		}()
	}
	return nil
}
