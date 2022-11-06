package core

import (
	"wa/xmpp"
)

func New_Hook_CallOffer(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		from, _ := n.GetAttr(`from`)
		id, _ := n.GetAttr(`id`)

		if len(n.Children) == 0 {
			return nil
		}
		ch0 := n.Children[0]
		if ch0.Tag != `offer` {
			return nil
		}

		ch0a := ch0.MapAttrs()
		call_creator, ok1 := ch0a[`call-creator`]
		call_id, ok2 := ch0a[`call-id`]
		if !ok1 || !ok2 {
			return nil
		}

		a.Noise.WriteXmppNode(&xmpp.Node{
			Tag: `receipt`,
			Attrs: []*xmpp.KeyValue{
				{Key: `id`, Value: id},
				{Key: `to`, Value: from},
			},
			Children: []*xmpp.Node{
				{
					Tag: ch0.Tag,
					Attrs: []*xmpp.KeyValue{
						{Key: `call-creator`, Value: call_creator},
						{Key: `call-id`, Value: call_id},
					},
				},
			},
		})

		return nil
	}
}

func New_Hook_CallAck(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		from, _ := n.GetAttr(`from`)
		id, _ := n.GetAttr(`id`)

		if len(n.Children) == 0 {
			return nil
		}
		ch0 := n.Children[0]

		switch ch0.Tag {
		case `relaylatency`:
		case `terminate`:
		default:
			return nil
		}

		a.Noise.WriteXmppNode(&xmpp.Node{
			Tag: `ack`,
			Attrs: []*xmpp.KeyValue{
				{Key: `class`, Value: `call`},
				{Key: `to`, Value: from},
				{Key: `id`, Value: id},
				{Key: `type`, Value: ch0.Tag},
			},
		})

		return nil
	}
}
