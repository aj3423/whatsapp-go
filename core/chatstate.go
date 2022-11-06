package core

import (
	"ajson"
	"wa/xmpp"
)

func (c Core) ChatState(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	jid := j.Get(`jid`).String()

	ch := &xmpp.Node{
		Tag: j.Get(`state`).String(),
	}
	if j.Get(`media_type`).String() == `ptt` {
		ch.SetAttr(`media`, `audio`)
	}

	e = a.Noise.WriteXmppNode(&xmpp.Node{
		Tag: `chatstate`,
		Attrs: []*xmpp.KeyValue{
			{Key: `to`, Value: jid, Type: 1},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}
