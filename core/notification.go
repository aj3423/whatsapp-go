package core

import (
	"wa/xmpp"
)

func New_Hook_Notification(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		type_, _ := n.GetAttr(`type`)
		from, _ := n.GetAttr(`from`)
		id, _ := n.GetAttr(`id`)
		participant, _ := n.GetAttr(`participant`)

		if type_ == `` || from == `` || id == `` {
			return nil
		}
		attrs := []*xmpp.KeyValue{
			{Key: `class`, Value: `notification`},
			{Key: `id`, Value: id},
			{Key: `to`, Value: from},
			{Key: `type`, Value: type_},
		}
		if len(participant) > 0 {
			attrs = append(attrs, &xmpp.KeyValue{
				Key: `participant`, Value: participant,
			})
		}
		a.Noise.WriteXmppNode(&xmpp.Node{
			Tag:   `ack`,
			Attrs: attrs,
		})

		return nil
	}
}
