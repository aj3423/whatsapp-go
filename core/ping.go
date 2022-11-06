package core

import (
	"event"
	"wa/xmpp"
)

func send_ping_ack(a *Acc) {
	a.Noise.WriteXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `result`},
		},
	})
}
func New_Hook_ServerPing(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		xmlns, _ := n.GetAttr(`xmlns`)

		if xmlns == "urn:xmpp:ping" {
			send_ping_ack(a)

			return event.Stop
		}

		return nil
	}
}

func New_Hook_Ping(a *Acc) func(...any) error {
	return func(...any) error {
		_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
			Tag: `iq`,
			Attrs: []*xmpp.KeyValue{
				{Key: `id`, Value: a.Noise.NextIqId_1()},
				{Key: `xmlns`, Value: `w:p`},
				{Key: `type`, Value: `get`},
				{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			},
			Children: []*xmpp.Node{
				{
					Tag: `ping`,
				},
			},
		})
		return e
	}
}
