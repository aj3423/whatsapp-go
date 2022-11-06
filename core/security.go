package core

import (
	"ajson"
	"wa/xmpp"

	"github.com/pkg/errors"
)

func (c Core) Set2fa(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	code, e := j.Get(`code`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`missing 'code'`))
	}
	email := j.Get(`email`).String()

	ch := []*xmpp.Node{
		{
			Tag:  `code`,
			Data: []byte(code),
		},
	}
	if code != `` {
		ch = append(ch, &xmpp.Node{
			Tag:  `email`,
			Data: []byte(email),
		})
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `urn:xmpp:whatsapp:account`},
		},
		Children: []*xmpp.Node{
			{
				Tag:      `2fa`,
				Children: ch,
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
