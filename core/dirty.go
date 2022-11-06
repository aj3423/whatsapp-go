package core

import (
	"wa/xmpp"
)

func (a *Acc) set_dirty_clean(type_ string) error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `urn:xmpp:whatsapp:dirty`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `clean`,
				Attrs: []*xmpp.KeyValue{
					{Key: `timestamp`, Value: `0`},
					{Key: `type`, Value: type_},
				},
			},
		},
	})
	return e
}
func New_Hook_Dirty(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		// if contains child `dirty`
		if ch, ok := n.FindChildByTag(`dirty`); ok {
			type_, _ := ch.GetAttr(`type`)

			if type_ == `groups` {
				a.Log.Warning(`groups dirty`)
				return a.Store.SetGroupsDirty()
			}
			if type_ == `account_sync` {

				a.Log.Warning(`account_sync dirty`)
				return a.Store.SetAccountSyncDirty()
			}
		}

		return nil
	}
}
