package core

import (
	"strconv"

	"ajson"
	"wa/xmpp"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const MaxMessageRetry = 0

const Receipt_Receive = 1 // tick 1
const Receipt_Read = 2    // tick 2

func build_receipt(
	id, to string,
	state int,
	is_group bool, participant string,
) []*xmpp.KeyValue {

	Attrs := []*xmpp.KeyValue{
		{Key: `to`, Value: to},
		{Key: `id`, Value: id},
	}
	if state == Receipt_Read {
		Attrs = append(Attrs,
			&xmpp.KeyValue{Key: `type`, Value: `read`})
	}
	if is_group {
		Attrs = append(Attrs,
			&xmpp.KeyValue{Key: `participant`, Value: participant},
		)
	}
	return Attrs
}

func (a *Acc) receipt(
	id, to string, state int,
	is_group bool, participant string,
) error {
	Attrs := build_receipt(id, to, state, is_group, participant)

	return a.Noise.WriteXmppNode(&xmpp.Node{
		Tag:   `receipt`,
		Attrs: Attrs,
	})
}
func (a *Acc) receipt_msg_receive(
	id, to string,
) error {
	return a.receipt(id, to, Receipt_Receive, false, ``)
}
func (a *Acc) receipt_group_msg_receive(
	id, to, participant string,
) error {
	return a.receipt(id, to, Receipt_Receive, true, participant)
}

func (c Core) ReceiptRead(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	id, e := j.Get(`id`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`missing 'id'`))
	}
	type_, e := j.Get(`type`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`missing 'type'`))
	}

	// read receipt multiple messages
	list, e_list := j.Get(`list`).TryStringArray()

	// load msg from db
	m, e := a.Store.GetMessage(id)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail get message, msg id: `+id))
	}
	n, e := xmpp.NewReader(m.Node).ReadNode()
	if e != nil {
		return NewErrRet(e)
	}
	ma := n.MapAttrs()

	from, ok := ma[`from`]
	if !ok {
		return NewErrRet(errors.New(`attrs missing from`))
	}
	participant, has_participant := ma[`participant`]

	attrs := []*xmpp.KeyValue{
		{Key: `id`, Value: id},
		{Key: `to`, Value: from},
		{Key: `type`, Value: type_},
	}
	if has_participant {
		attrs = append(attrs, &xmpp.KeyValue{
			Key: `participant`, Value: participant})
	}

	nreq := &xmpp.Node{
		Tag:   `receipt`,
		Attrs: attrs,
	}

	if e_list == nil { // multi messages
		ch := []*xmpp.Node{}
		for _, id := range list {
			ch = append(ch, &xmpp.Node{
				Tag: `item`,
				Attrs: []*xmpp.KeyValue{
					{Key: `id`, Value: id},
				},
			})
		}
		nreq.Children = []*xmpp.Node{
			{
				Tag:      `list`,
				Children: ch,
			},
		}
	}
	nr, e := a.Noise.WriteReadXmppNode(nreq)
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}

func (a *Acc) retry_receipt(
	attrs map[string]string,
	direct_distribution bool,
) error {
	id, ok1 := attrs[`id`]
	from, ok2 := attrs[`from`]
	t, ok3 := attrs[`t`]
	if !ok1 || !ok2 || !ok3 {
		return errors.New(`no attr`)
	}

	n_registration, e1 := a.registration_node()
	n_prekey, e2 := a.prekey_node()
	n_spk, e3 := a.spk_node()
	n_identity, e4 := a.identity_node()
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return multierr.Combine(e1, e2, e3, e4)
	}

	cnt := 1
	if direct_distribution {
		cnt += 1
	}
	n := &xmpp.Node{
		Tag: `receipt`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: id},
			{Key: `to`, Value: from},
			{Key: `type`, Value: `retry`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `retry`,
				Attrs: []*xmpp.KeyValue{
					{Key: `count`, Value: strconv.Itoa(cnt)},
					{Key: `id`, Value: id},
					{Key: `t`, Value: t},
					{Key: `v`, Value: `1`}, // `1` for both group/wm
				},
			},
			n_registration,
		},
	}
	if participant, ok := attrs[`participant`]; ok {
		n.Attrs = append(n.Attrs, &xmpp.KeyValue{
			Key: `participant`, Value: participant})
	}
	if direct_distribution {
		n.Children = append(n.Children, &xmpp.Node{
			Tag: `keys`,
			Children: []*xmpp.Node{
				n_identity,
				a.djb_type_node(),
				n_prekey,
				n_spk,
			},
		})
	}

	a.Log.Info("retry msg: %s", id)

	_, e := a.Noise.WriteReadXmppNode(n)

	if e == nil {
		a.Store.IncreaseMessageRetry(id)
	}
	return e
}

func New_Hook_Receipt(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		ma := n.MapAttrs()

		// all ack have `id`, `to`
		id, _ := ma[`id`]
		from, _ := ma[`from`]

		// group has `participant`
		participant, _ := ma[`participant`]

		// type == `Receive` or `Read`
		type_, _ := ma[`type`]

		Attrs := []*xmpp.KeyValue{
			{Key: `class`, Value: `receipt`},
			{Key: `id`, Value: id},
			{Key: `to`, Value: from},
		}
		if len(participant) > 0 {
			Attrs = append(Attrs, &xmpp.KeyValue{Key: `participant`, Value: participant})
		}
		if len(type_) > 0 {
			Attrs = append(Attrs, &xmpp.KeyValue{Key: `type`, Value: type_})
		}
		return a.Noise.WriteXmppNode(&xmpp.Node{
			Tag:   `ack`,
			Attrs: Attrs,
		})
	}
}
