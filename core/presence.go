package core

import (
	"ajson"
	"wa/xmpp"
)

func (a *Acc) presence(type_, nick, jid string) error {
	// type == `available`/`unavailable`/`subscribe`
	Attrs := []*xmpp.KeyValue{
		{Key: `type`, Value: type_},
	}

	if type_ == `subscribe` {
		Attrs = append(Attrs, &xmpp.KeyValue{
			Key: `to`, Value: jid,
		})
	}
	// has nick
	if len(nick) > 0 {
		Attrs = append(Attrs, &xmpp.KeyValue{
			Key: `name`, Value: nick,
		})
	}

	return a.Noise.WriteXmppNode(&xmpp.Node{
		Tag:   `presence`,
		Attrs: Attrs,
	})
}

func (c Core) Presence(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	type_ := j.Get(`type`).String()
	jid := j.Get(`jid`).String()
	nick := j.Get(`nick`).String()
	e = a.presence(type_, nick, jid)
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}
