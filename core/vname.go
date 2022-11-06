package core

import (
	"ahex"
	"ajson"
	"algo/xed25519"
	"arand"
	"wa/pb"
	"wa/xmpp"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (a *Acc) load_vname(include_nick, include_sig bool) (*pb.VName, error) {
	prof, e := a.Store.GetProfile()
	if e != nil {
		return nil, errors.Wrap(e, "fail get profile")
	}

	cfg, e := a.Store.GetConfig()
	if e != nil {
		return nil, errors.Wrap(e, "fail get config")
	}
	vname := &pb.VName{}
	e = proto.Unmarshal(cfg.VNameCert, vname)
	if e != nil {
		return nil, errors.Wrap(e, `wrong cfg.VNameCert: `+ahex.Enc(cfg.VNameCert))
	}
	if include_nick {
		if vname.Cert == nil {
			vname.Cert = &pb.VName_CERT{}
		}
		vname.Cert.Nick = proto.String(prof.Nick)
	}

	if include_sig {
		e = a.sign_vname(vname)
		if e != nil {
			return nil, errors.Wrap(e, "fail sign vname")
		}
	}
	return vname, nil
}

// only for Register()
func (a *Acc) generate_vname() (*pb.VName, error) {
	vname := &pb.VName{
		Cert: &pb.VName_CERT{
			Rand_64: proto.Uint64(uint64(arand.Int(0x1000000000000000, 0x7fffffffffffffff))),
			SmbWa:   proto.String("smb:wa"),
			Nick:    proto.String(""),
		},
	}
	e := a.sign_vname(vname)
	return vname, e
}
func (a *Acc) sign_vname(vname *pb.VName) error {
	bs, _ := proto.Marshal(vname.Cert)

	iden, e := a.Store.GetMyIdentity()
	if e != nil {
		return e
	}
	vname.Signature, e = xed25519.Sign(iden.PrivateKey, bs)
	return e
}

func (a *Acc) set_biz_verified_name() error {
	vname, e := a.load_vname(true, true)
	if e != nil {
		return e
	}

	bs, _ := proto.Marshal(vname)

	_, e = a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:biz`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `verified_name`,
				Attrs: []*xmpp.KeyValue{
					{Key: `v`, Value: `2`},
				},
				Data: bs,
			},
		},
	})
	return e
}
func (c Core) SetBizVerifiedName(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	// 1. save to db
	rj := c.SetMyProfile(j)
	if rj.Get(`ErrCode`).Int() != 0 {
		return rj
	}

	// 2. commit
	e = a.set_biz_verified_name()
	if e != nil {
		return NewErrRet(e)
	}
	return NewSucc()
}

func (a *Acc) get_biz_verified_name() error {
	dev, e := a.Store.GetDev()
	if e != nil {
		return e
	}

	_, e = a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:biz`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `verified_name`,
				Attrs: []*xmpp.KeyValue{
					{Key: `jid`, Value: dev.Cc + dev.Phone + "@s.whatsapp.net"},
				},
			},
		},
	})
	return e
}
func (c Core) GetBizVerifiedName(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	e = a.get_biz_verified_name()
	if e != nil {
		return NewErrRet(e)
	}
	return NewSucc()
}
