package core

import (
	"encoding/json"
	"fmt"

	"ajson"
	"wa/crypto"
	"wa/signal/protocol"
	"wa/xmpp"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const Encrypt_Upload_Size = 812

type EncryptResult struct {
	Jid      string
	RegId    uint32
	Identity []byte
	SpkId    uint32
	Spk      []byte
	SpkSig   []byte
	Prekey   []byte
	PrekeyId uint32
}

/*
	wa returns
        "code": "406", "text": "not-acceptable"
	if the device doesn't exist, eg: xxx.0:999@s.whatsapp.net
*/

func (a *Acc) get_encrypt(
	jids []string, tag string,
) (map[string]EncryptResult, error) {
	ret := map[string]EncryptResult{}

	ch := []*xmpp.Node{}
	for _, jid := range jids {
		ch = append(ch, &xmpp.Node{
			Tag: `user`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
		})
	}
	n := &xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `encrypt`},
		},
		Children: []*xmpp.Node{
			{
				Tag:      tag, // `key`/`identity`
				Children: ch,
			},
		},
	}

	rn, e := a.Noise.WriteReadXmppNode(n)
	if e != nil {
		return nil, e
	}

	if rn.ToJson().Get(`Attrs`).Get(`type`).String() != `result` {
		return nil, errors.New(`fail get encrypt: ` + rn.ToString())
	}
	for _, ch := range rn.Children {
		if ch.Tag != `list` {
			continue
		}

		for _, user := range ch.Children {
			jid, _ := user.GetAttr(`jid`)
			er := EncryptResult{Jid: jid}

			for _, u_ch := range user.Children {
				switch u_ch.Tag {
				case `registration`:
					er.RegId = crypto.BE2U32(u_ch.Data)
				case `identity`:
					er.Identity = u_ch.Data
				case `skey`:
					for _, jj := range u_ch.Children {
						switch jj.Tag {
						case `id`:
							er.SpkId = crypto.BE2U24(jj.Data)
						case `value`:
							er.Spk = jj.Data
						case `signature`:
							er.SpkSig = jj.Data
						}
					}
				case `key`:
					for _, jj := range u_ch.Children {
						switch jj.Tag {
						case `id`:
							er.PrekeyId = crypto.BE2U24(jj.Data)
						case `value`:
							er.Prekey = jj.Data
						}
					}
				}
			}

			ret[jid] = er
		}
	}

	return ret, nil
}

func (c Core) GetEncrypt(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	jids, e := j.Get(`jids`).TryStringArray()
	if e != nil {
		return NewErrRet(errors.New(`invalid param "jids"`))
	}

	tag, e := j.Get(`tag`).TryString()

	if e != nil {
		return NewErrRet(errors.New(`invalid param "tag"`))
	}

	emap, e := a.get_encrypt(jids, tag)
	if e != nil {
		return NewErrRet(e)
	}

	// convert map -> json -> ajson
	bs, e := json.Marshal(emap)
	if e != nil {
		return NewErrRet(e)
	}
	rj, e := ajson.ParseByte(bs)
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(rj)
}
func (a *Acc) set_encrypt() error {
	n_identity, e1 := a.identity_node()
	n_registration, e2 := a.registration_node()
	n_spk, e3 := a.spk_node()
	n_prekey, e4 := a.prekeys_node()

	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return multierr.Combine(e1, e2, e3, e4)
	}

	n := &xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `encrypt`},
		},
		Children: []*xmpp.Node{
			n_identity,
			n_registration,
			a.djb_type_node(),
			n_prekey,
			n_spk,
		},
	}

	_, e := a.Noise.WriteReadXmppNode(n)
	return e
}

// not enough prekeys on server, need to upload more
func New_Hook_NeedMorePrekeys(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		type_, _ := attrs[`type`]
		if type_ != `encrypt` {
			return nil
		}

		from, _ := attrs[`from`]
		if from != `s.whatsapp.net` {
			return nil
		}

		_, ok := n.FindChildByTag(`count`)
		if !ok {
			return nil
		}

		return a.set_encrypt()
	}
}

// someone re-registered, identity change
// clear identity/session of the jid
func New_Hook_PeerIdentityChange(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		type_, _ := attrs[`type`]
		if type_ != `encrypt` {
			return nil
		}

		from, ok := attrs[`from`]
		if !ok {
			return nil
		}

		_, ok = n.FindChildByTag(`identity`)
		if !ok {
			return nil
		}
		recid, devid, e := split_jid(from)
		if e != nil {
			return e
		}
		addr := protocol.NewSignalAddress(fmt.Sprintf("%d", recid), devid)

		a.Store.DeleteIdentity(addr)
		a.Store.DeleteSession(addr)
		a.Store.DeleteSenderKey(addr)

		return nil
	}
}

func (a *Acc) identity_node() (*xmpp.Node, error) {
	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return nil, e
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	return &xmpp.Node{
		Tag:  `identity`,
		Data: iden_pub[:],
	}, nil
}
func (a *Acc) registration_node() (*xmpp.Node, error) {
	dev, e := a.Store.GetDev()
	if e != nil {
		return nil, e
	}
	return &xmpp.Node{
		Tag:  `registration`,
		Data: crypto.U322BE(dev.RegId),
	}, nil
}
func (a *Acc) spk_node() (*xmpp.Node, error) {
	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return nil, e
	}
	spk_key := spk.KeyPair().PublicKey().PublicKey()
	spk_sig := spk.Signature()
	return &xmpp.Node{
		Tag: `skey`,
		Children: []*xmpp.Node{
			{Tag: `id`, Data: []byte{0, 0, 0}}, // prekey_id
			{Tag: `value`, Data: spk_key[:]},
			{Tag: `signature`, Data: spk_sig[:]},
		},
	}, nil
}
func (a *Acc) djb_type_node() *xmpp.Node {
	return &xmpp.Node{
		Tag:  `type`,
		Data: []byte{5},
	} // djbType
}
func (a *Acc) prekeys_node() (*xmpp.Node, error) {
	prekeys := []*xmpp.Node{}
	for i := 0; i < Encrypt_Upload_Size; i++ {
		kid, e := a.Store.GetMyNextPrekeyId()
		if e != nil {
			return nil, e
		}
		rec, e := a.Store.GeneratePrekey(kid)
		if e != nil {
			return nil, e
		}
		pub := rec.KeyPair().PublicKey().PublicKey()

		prekeys = append(prekeys, &xmpp.Node{
			Tag: `key`,
			Children: []*xmpp.Node{
				{Tag: `id`, Data: crypto.U322BE_24(kid)},
				{Tag: `value`, Data: pub[:]},
			},
		})
	}
	return &xmpp.Node{ // prekeys
		Tag:      `list`,
		Children: prekeys,
	}, nil
}
func (a *Acc) prekey_node() (*xmpp.Node, error) {
	kid, e := a.Store.GetMyNextPrekeyId()
	if e != nil {
		return nil, e
	}
	rec, e := a.Store.GeneratePrekey(kid)
	if e != nil {
		return nil, e
	}
	pub := rec.KeyPair().PublicKey().PublicKey()

	return &xmpp.Node{
		Tag: `key`,
		Children: []*xmpp.Node{
			{Tag: `id`, Data: crypto.U322BE_24(kid)},
			{Tag: `value`, Data: pub[:]},
		},
	}, nil
}
