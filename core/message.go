package core

import (
	"fmt"
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
	"wa/signal/protocol"
	"wa/signal/session"
	"wa/xmpp"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const Msg_Cipher_Ver = `2`

// protocol.WHISPER_TYPE                = 2
// protocol.PREKEY_TYPE                 = 3
// protocol.SENDERKEY_TYPE              = 4
// protocol.SENDERKEY_DISTRIBUTION_TYPE = 5
func msg_type_str(t uint32) string {
	switch t {
	case protocol.WHISPER_TYPE:
		return `msg`
	case protocol.PREKEY_TYPE:
		return `pkmsg`
	case protocol.SENDERKEY_TYPE:
		return `skmsg`
	case protocol.SENDERKEY_DISTRIBUTION_TYPE:
		panic("unimplemented msg type skdm")
	}
	panic(fmt.Sprintf("unimplemented wtf type: %d", t))
}
func msg_type_int(t string) uint32 {
	switch t {
	case `msg`:
		return protocol.WHISPER_TYPE
	case `pkmsg`:
		return protocol.PREKEY_TYPE
	case `skmsg`:
		return protocol.SENDERKEY_TYPE
	}
	panic(fmt.Sprintf("unimplemented wtf type: %s", t))
}

type MessageContent struct {
	P *pb.Message
}

func NewMessageContentFromDecrypted(
	media Media,
) *MessageContent {
	ret := &MessageContent{P: &pb.Message{}}

	switch media.Type() {
	case pb.Media_Text:
		ret.P.Text = media.Serialize()
	case pb.Media_Image:
		img := &pb.Image{}
		if e := proto.Unmarshal(media.Serialize(), img); e == nil {
			ret.P.Image = img
		}
	case pb.Media_Ptt:
		ptt := &pb.Ptt{}
		if e := proto.Unmarshal(media.Serialize(), ptt); e == nil {
			ret.P.Ptt = ptt
		}
	case pb.Media_Sticker:
		emj := &pb.Sticker{}
		if e := proto.Unmarshal(media.Serialize(), emj); e == nil {
			ret.P.Sticker = emj
		}
	case pb.Media_Video:
		vid := &pb.Video{}
		if e := proto.Unmarshal(media.Serialize(), vid); e == nil {
			ret.P.Video = vid
		}
	case pb.Media_Url:
		bc := &pb.Url{}
		if e := proto.Unmarshal(media.Serialize(), bc); e == nil {
			ret.P.Url = bc
		}
	case pb.Media_Document:
		bc := &pb.Document{}
		if e := proto.Unmarshal(media.Serialize(), bc); e == nil {
			ret.P.Document = bc
		}
	}

	return ret
}
func (mc *MessageContent) GetMedia() (
	media Media,
) {
	if mc.P.Text != nil {
		return &Text{Data: mc.P.Text}
	}
	if mc.P.Image != nil {
		return &Image{P: mc.P.Image}
	}
	if mc.P.Video != nil {
		return &Video{P: mc.P.Video}
	}
	if mc.P.Ptt != nil {
		return &Ptt{P: mc.P.Ptt}
	}
	if mc.P.Sticker != nil {
		return &Sticker{P: mc.P.Sticker}
	}
	if mc.P.Url != nil {
		return &Url{P: mc.P.Url}
	}
	if mc.P.Document != nil {
		return &Document{P: mc.P.Document}
	}

	return &TodoMedia{}
}

func (mc *MessageContent) Skdm() (*protocol.SenderKeyDistributionMessage, error) {
	skdm_bytes := mc.P.GetGrp().GetSkdm()
	return protocol.NewSenderKeyDistributionMessageFromBytes(skdm_bytes)
}

/*
[ "111@whatsapp.net", "222@whatsapp.net"]
->
[ "111@whatsapp.net", "111.0:2@whatsapp.net", "222@whatsapp.net", ...]
*/
func expand_jids_devices(a *Acc, jids []string) ([]string, error) {
	ret := []string{}
	for _, jid := range jids {
		recid, _, e := split_jid(jid)
		if e != nil {
			return nil, e
		}
		devids, e := a.Store.GetMultiDevice(recid)
		if e != nil {
			return nil, e
		}

		// append phone as main device without .0:0
		ret = append(ret, fmt.Sprintf("%d@s.whatsapp.net", recid))

		for _, devid := range devids {
			ret = append(ret, fmt.Sprintf("%d.0:%d@s.whatsapp.net", recid, devid))
		}
	}
	return ret, nil
}

/*
build an array of:

	{
	  "Attrs": {
		"jid": "8618311112222.0:3@s.whatsapp.net"
	  },
	  "Children": [
		{
		  "Attrs": {
			"type": "pkmsg",
			"v": "2"
		  },
		  "Data": "33084d122105e4b17ac7d015e32be01a9499c6b6fadaeef8059e7ddc1e6ff5d739019abad12e1a2105123bb7c516a566706339c7b81b021055051992512eda51e9c7ac292bcdbad6332262330a2105a04169bff1c64f06b6f2e75d427e1f22c6054ff445738d4cc72831432b6c3c0110001800223086339ce433a08b84479e0aa54ad41c8425035b00fcb6953bbb0036a7cb00cf77af1bb95ff583357359c0aee14349f71074a70e4ac19a0e9028bc97d8ca073001",
		  "Tag": "enc"
		}
	  ],
	  "Tag": "to"
	}
*/
func build_participants_node(
	a *Acc, jids []string,
	padded_msg []byte,
	media_t pb.Media_Type,
	media Media,
) ([]*xmpp.Node, error) {
	ret := []*xmpp.Node{}

	// get all expanded jids, including multi-device
	// eg: [`111@s.whatsapp.net`, `111.0:2@s.whatsapp.net`, `222@...`]
	jids, e := expand_jids_devices(a, jids)
	if e != nil {
		return nil, e
	}

	// create session if not exists
	map_sb, e := a.ensure_session_builder(jids)
	if e != nil {
		return nil, e
	}

	for _, jid := range jids {
		recid, devid, _ := split_jid(jid)

		peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)

		sb := map_sb[jid] // guaranteed exists

		sc := session.NewCipher(sb, peer_addr)

		// encMsg is
		// 33 08 22 12 ....
		encMsg, e := sc.Encrypt(padded_msg)
		if e != nil {
			return nil, e
		}

		Attrs := []*xmpp.KeyValue{
			{Key: `v`, Value: Msg_Cipher_Ver},
			{Key: `type`, Value: msg_type_str(encMsg.Type())},
		}
		// `ptt`/`...`
		if media_t != pb.Media_Text {
			Attrs = append(Attrs, &xmpp.KeyValue{
				Key: `mediatype`, Value: MediaTypeStr(media.Type()),
			})
		}

		ret = append(ret, &xmpp.Node{
			Tag: `to`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
			Children: []*xmpp.Node{
				{
					Tag:   `enc`,
					Attrs: Attrs,
					Data:  encMsg.Serialize(),
				},
			},
		})
	}
	return ret, nil
}

func (c Core) SendMsg(j *ajson.Json) *ajson.Json {
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
	pmsg := &pb.Message{}
	media.FillMessage(pmsg)

	p, _ := proto.Marshal(pmsg)
	padded := crypto.RandomPadMsg(p)

	// build session
	jid := j.Get(`jid`).String()

	ptcps, e := build_participants_node(a, []string{jid}, padded, media_t, media)
	if e != nil || len(ptcps) == 0 {
		return NewErrRet(errors.New("wrong jid: " + jid))
	}

	// WamE2eMessageSend
	{
		enc_msg_type, _ := ptcps[0].GetAttr(`type`) // only log the first participant
		e = a.wam_e2e_message_send(
			media_t,
			wam_cipher_text_type(enc_msg_type),
			wam_cipher_text_ver(Msg_Cipher_Ver),
			0, /*dest*/
		)
		if e != nil {
			a.Log.Error(`fail wam_e2e_message_send: ` + e.Error())
		}
	}

	n := &xmpp.Node{
		Tag: `message`,
		Attrs: []*xmpp.KeyValue{
			{Key: `to`, Value: jid, Type: 1},
			{Key: `type`, Value: media.MsgCategory()},
			{Key: `id`, Value: strings.ToUpper(algo.Md5Str([]byte(arand.Uuid4())))},
		},
	}

	if len(ptcps) == 1 { // only phone, only 1 child node
		n.Children = []*xmpp.Node{ptcps[0].Children[0]}
	} else { // > 1, multi device
		n.Children = []*xmpp.Node{
			{
				Tag:      `participants`,
				Children: ptcps,
			},
		}
	}

	if j.Exists(`url_number`) { // sending msg from clicking url `https://api.whatsapp.com/send?phone=...`
		n.Children = append([]*xmpp.Node{
			{Tag: `url_number`},
		}, n.Children...)
	}

	// send msg
	nr, e := a.Noise.WriteReadXmppNode(n)

	if e != nil {
		return NewErrRet(e)
	}

	// only save on success, don't know the failure value for messageSendResult
	{
		dev, e := a.Store.GetDev()
		if e != nil {
			return NewErrRet(e)
		}
		international := !strings.HasPrefix(jid, dev.Cc)
		msg_type := 1 // 1: personal
		send_begin := time.Now()
		er := a.wam_message_send(media_t, send_begin, msg_type, international)
		if er != nil {
			a.Log.Error(`fail WamMessageSend: ` + er.Error())
		}
	}

	return NewJsonRet(nr.ToJson())
}

func (a *Acc) cipherFromJid(
	jid string,
) *session.Cipher {
	recid, devid, _ := split_jid(jid)

	peer_addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)
	sb := session.NewBuilder(
		a.Store, a.Store, a.Store, a.Store, peer_addr)
	sc := session.NewCipher(sb, peer_addr)
	return sc
}

func (a *Acc) decodeWspProtoFromJid(
	jid string,
	cipher []byte,
	msg_type uint32,
) ([]byte, error) {
	sc := a.cipherFromJid(jid)

	var bs []byte
	var e error
	switch msg_type {
	case protocol.WHISPER_TYPE:
		bs, e = decodeWspMsg(sc, cipher)
	case protocol.PREKEY_TYPE:
		bs, e = decodePkMsg(sc, cipher)
	}
	if e != nil {
		return nil, e
	}

	return crypto.UnPadMsg(bs)
}
func parseMessageProto(
	unpadded_proto []byte,
) (
	*MessageContent, error,
) {
	// parse proto
	m := &pb.Message{}
	e := proto.Unmarshal(unpadded_proto, m)
	if e != nil {
		return nil, e
	}
	return &MessageContent{P: m}, nil
}
func (a *Acc) decodeWspMsgTextFromJid(
	jid string,
	cipher []byte,
	msg_type uint32,
) (
	*MessageContent, error,
) {
	unpadded, e := a.decodeWspProtoFromJid(
		jid, cipher, msg_type)
	if e != nil {
		return nil, e
	}

	return parseMessageProto(unpadded)
}

func decodeWspMsg(sc *session.Cipher, cipher []byte) ([]byte, error) {
	// whisper msg
	wm, e := protocol.NewSignalMessageFromBytes(cipher)
	if e != nil {
		return nil, e
	}
	return sc.Decrypt(wm)
}
func decodePkMsg(sc *session.Cipher, cipher []byte) ([]byte, error) {
	pkm, e := protocol.NewPreKeySignalMessageFromBytes(cipher)
	if e != nil {
		return nil, e
	}

	return sc.DecryptMessage(pkm)
}

func New_Hook_Msg(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		// return if group msg (has attr `participant`)
		{
			if _, ok := attrs[`participant`]; ok {
				return nil
			}
		}

		// must exists
		msg_id, ok1 := attrs[`id`]
		from, ok2 := attrs[`from`]
		t, ok3 := attrs[`t`]

		if !ok1 || !ok2 || !ok3 {
			return nil
		}

		var msg_type string
		wmNode, ok := n.FindChild(func(ch *xmpp.Node) bool {
			msg_type, _ = ch.GetAttr(`type`)
			return ch.Tag == `enc` && (msg_type == `pkmsg` || msg_type == `msg`)
		})
		if !ok {
			return nil
		}

		// usync multi-device for any new peer account
		defer func() {
			a.Wg.Add(1)
			// must use another goroutine, otherwise tcp would be blocking
			// usync() would freeze for 1 minute
			go func() {
				defer phoenix.Ignore(nil)
				defer a.Wg.Done()

				recid, devid, e := split_jid(from)
				if e != nil {
					return
				}

				// check last sync time
				t, e := a.Store.GetMultiDeviceLastSync(recid, devid)
				if e != nil {
					return
				}

				if time.Since(t) < 30*24*time.Hour { // synced recently
					return
				}

				nr, e := a.usync_multi_device(
					fmt.Sprintf("%d@s.whatsapp.net", recid))
				if e != nil {
					return
				}
				e = a.handle_usync_multi_device_result(nr, recid)
				if e != nil {
					return
				}
				a.Store.SetMultiDeviceLastSync(recid, devid)
			}()
		}()

		// 0. event
		ev := event.New[string]()
		ev.On(def.Ev_Retry, func(...any) error {
			if e := a.retry_message(msg_id, attrs); e != nil {
				// error means too much retry
				// just receipt or connection'll be lost
				return a.receipt_msg_receive(msg_id, from)
			}
			return event.Stop
		})
		ev.On(def.Ev_Success, func(args ...any) error {
			//  send receipt
			{
				a.receipt_msg_receive(msg_id, from)
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
				cipher_type := attrs[`type`]
				cipher_ver := attrs[`v`]

				timestamp, _ := strconv.Atoi(t)
				international := !strings.HasPrefix(from, dev.Cc)

				a.wam_e2e_message_recv(
					media.Type(),
					timestamp,
					wam_cipher_text_type(cipher_type),
					wam_cipher_text_ver(cipher_ver),
					0, /*dest*/
				)

				a.wam_message_receive(
					media.Type(),
					timestamp,
					1, /*msg_type*/
					international)
			}

			// 3. modify node
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

			// 5. store decoded to db
			return a.Store.SaveDecodedMessage(
				msg_id, uint32(media.Type()), media.Serialize())
		})

		// 1. save to database
		{
			m, e := a.Store.EnsureMessage(
				msg_id, xmpp.NewWriter().WriteNode(n))
			if e != nil {
				return e
			}
			// if already decrypt,
			//   send receipt and push to client
			if m.Decrypted {
				a.Log.Error("msg already decrypted: %s", msg_id)

				med, e := NewMediaFromBytes(
					pb.Media_Type(m.MediaType), m.DecMedia)
				if e != nil {
					return e
				}
				return ev.Fire(`success`, NewMessageContentFromDecrypted(
					med,
				))
			}
		}

		// 2. decrypt
		var e error
		var content *MessageContent

		switch msg_type {
		case `msg`:
			content, e = a.decodeWspMsgTextFromJid(
				from, wmNode.Data, protocol.WHISPER_TYPE)
		case `pkmsg`:
			content, e = a.decodeWspMsgTextFromJid(
				from, wmNode.Data, protocol.PREKEY_TYPE)
		default:
			return errors.New("wtf msg_type: " + msg_type)
		}

		// retry on decode error
		if e != nil {
			// send `retry` if any error occurs
			a.Log.Warning("decode msg fail: %s, retry", e.Error())
			return ev.Fire(`retry`)
		}

		return ev.Fire(`success`, content)
	}
}
func (a *Acc) retry_message(
	msg_id string,
	attrs map[string]string,
) error {
	retry_times, e := a.Store.GetMessageRetry(msg_id)
	if e != nil {
		a.Log.Error("fail Get Retry times for msg %s: %s", msg_id, e.Error())
		return e
	}
	if retry_times >= MaxMessageRetry {
		errStr := fmt.Sprintf("too much retry for msg %s", msg_id)
		a.Log.Warning(errStr)
		return errors.New(errStr)
	}

	{ // retry
		a.Wg.Add(1)
		go func() {
			defer phoenix.Ignore(nil)

			a.retry_receipt(attrs, false)
			a.Wg.Done()
		}()
	}

	return nil
}
func (c Core) ListMsgCache(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	msgs, e := a.Store.ListMessages()
	if e != nil {
		return NewErrRet(e)
	}

	var ret_arr []*ajson.Json
	for _, m := range msgs {
		med, e := NewMediaFromBytes(
			pb.Media_Type(m.MediaType), m.DecMedia)
		if e != nil {
			return NewErrRet(e)
		}

		x := med.ToJson()
		x.Set(`id`, m.MsgId)
		x.Set(`decrypted`, m.Decrypted)
		x.Set(`updated_at`, m.UpdatedAt.Unix())
		ret_arr = append(ret_arr, x)
	}

	r := NewSucc()
	r.Set(`messages`, ret_arr)
	return r
}
func (c Core) DeleteMsgCache(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	ids, e := j.Get(`ids`).TryStringArray()
	if e != nil {
		return NewErrRet(errors.New(`invalid param 'ids'`))
	}

	e = a.Store.DeleteMessages(ids)
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}
