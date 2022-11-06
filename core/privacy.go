package core

import (
	"ajson"
	"wa/xmpp"
)

func (a *Acc) get_jabber_iq_privacy() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `jabber:iq:privacy`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `query`,
				Children: []*xmpp.Node{
					{
						Tag: `list`,
						Attrs: []*xmpp.KeyValue{
							{Key: `name`, Value: `default`},
						},
					},
				},
			},
		},
	})
	return e
}
func (a *Acc) get_status_privacy() error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `status`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `privacy`,
			},
		},
	})
	return e
}
func (a *Acc) get_privacy(xmlns string) error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: xmlns},
		},
		Children: []*xmpp.Node{
			{
				Tag: `privacy`,
			},
		},
	})
	return e
}

func (a *Acc) set_status_privacy() (*xmpp.Node, error) {
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `status`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `privacy`,
				Children: []*xmpp.Node{
					{
						Tag: `list`,
						Attrs: []*xmpp.KeyValue{
							{Key: `type`, Value: `contacts`},
						},
					},
				},
			},
		},
	})
	return nr, e
}

func (c Core) SetStatusPrivacy(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	nr, e := a.set_status_privacy()
	if e != nil {
		return NewErrRet(e)
	}
	return NewJsonRet(nr.ToJson())
}
