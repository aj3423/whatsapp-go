package core

import (
	"strings"

	"ajson"
	"algo"
	"arand"
	"wa/xmpp"

	"github.com/pkg/errors"
)

func (c Core) CreateGroup(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	subject := j.Get("subject").String()

	participants, e := j.Get("participants").TryStringArray()
	if e != nil {
		return NewErrRet(errors.New(`wrong param participants`))
	}

	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	recid := dev.Cc + dev.Phone

	// participants
	ptcps := &xmpp.Node{
		Tag: `create`,
		Attrs: []*xmpp.KeyValue{
			{Key: `key`, Value: recid + `-` + algo.Md5Str([]byte(arand.Uuid4())) + `@temp`},
			{Key: `subject`, Value: subject},
		},
	}
	for _, jid := range participants {
		child := &xmpp.Node{
			Tag: `participant`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
		}
		ptcps.Children = append(ptcps.Children, child)
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `g.us`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{ptcps},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) QueryGroup(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	ch := &xmpp.Node{
		Tag: `query`,
	}
	if request, err := j.Get(`request`).TryString(); err == nil {
		if ch.Attrs == nil {
			ch.Attrs = []*xmpp.KeyValue{}
		}
		ch.Attrs = append(ch.Attrs, &xmpp.KeyValue{
			Key: `request`, Value: request,
		})
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) GroupDesc(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	body, e := j.Get("body").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param body`))
	}

	prev_id := j.Get("prev_id").String()

	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	recid := dev.Cc + dev.Phone

	/*
		jadx for "unable to provide message id hash due to missing md5 algorithm"
		md5(
			Long.valueOf(System.currentTimeMillis()) or SystemClock.elapsedRealtime(),
			jid,
			SecureRandom.nextBytes(16),
		)
	*/
	id := strings.ToUpper(algo.Md5Str([]byte(
		recid + arand.Uuid4(), // just for simple
	)))

	ch := &xmpp.Node{
		Tag: `description`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: id},
		},
		Children: []*xmpp.Node{
			{Tag: `body`, Data: []byte(body)},
		},
	}
	if len(prev_id) > 0 {
		ch.Attrs = append(ch.Attrs, &xmpp.KeyValue{
			Key: `prev`, Value: prev_id,
		})
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) LeaveGroup(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `g.us`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `leave`,
				Children: []*xmpp.Node{
					{
						Tag: `group`,
						Attrs: []*xmpp.KeyValue{
							{Key: `id`, Value: gid},
						},
					},
				},
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) AddGroupMember(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	jids, e := j.Get("jids").TryStringArray()
	if e != nil {
		return NewErrRet(errors.New(`wrong param jids, not string array`))
	}
	to_add := []*xmpp.Node{}
	for _, jid := range jids {
		to_add = append(to_add, &xmpp.Node{
			Tag: `participant`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
		})
	}
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_1()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			{
				Tag:      `add`,
				Children: to_add,
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) RemoveGroupMember(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	jid, e := j.Get("jid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param jid`))
	}
	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `remove`,
				Children: []*xmpp.Node{
					{
						Tag: `participant`,
						Attrs: []*xmpp.KeyValue{
							{Key: `jid`, Value: jid},
						},
					},
				},
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) ListGroup(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:g2`},
			{Key: `to`, Value: `g.us`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `participating`,
				Children: []*xmpp.Node{
					{Tag: `participants`},
					{Tag: `description`},
				},
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	if chGroups, ok := nr.FindChildByTag(`groups`); ok {
		// 1. clear
		e1 := a.Store.RemoveAllGroups()
		e2 := a.Store.RemoveAllGroupMembers()
		if e1 != nil || e2 != nil {
			return NewErrRet(errors.New(`fail update group info`))
		}
		// 2. update
		for _, g := range chGroups.Children {
			attrs := g.MapAttrs()

			id, ok1 := attrs[`id`]
			subject, ok2 := attrs[`subject`]
			if !ok1 || !ok2 {
				continue
			}

			// creator can be empty
			creator := attrs[`creator`]

			var members []string

			for _, mbr := range g.Children {
				switch mbr.Tag {
				case `description`:
					continue
				case `participant`:
					mbr_attrs := mbr.MapAttrs()
					jid, ok := mbr_attrs[`jid`]
					if ok {
						members = append(members, jid)
					}
				}
			}

			if e := a.Store.CreateGroup(
				id+`@g.us`, subject, creator, members,
			); e != nil {
				continue
			}
		}
	}

	return NewJsonRet(nr.ToJson())
}

func New_Hook_GroupCreate(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		// MUST have attr `from`,`participant`
		from, ok1 := attrs[`from`]
		type_, ok2 := attrs[`type`]

		if !ok1 || !ok2 {
			return nil
		}
		if type_ != `w:gp2` {
			return nil
		}
		chCreate, ok := n.FindChildByTag(`create`)
		if !ok {
			return nil
		}
		chGroup, ok := chCreate.FindChildByTag(`group`)
		if !ok {
			return nil
		}
		grpAttrs := chGroup.MapAttrs()
		creator, ok1 := grpAttrs[`creator`]
		subject, ok2 := grpAttrs[`subject`]
		if !ok1 || !ok2 {
			return nil
		}

		members := []string{}
		for _, ch := range chGroup.Children {
			if ch.Tag == `participant` {
				m := ch.MapAttrs()
				jid, ok := m[`jid`]
				if !ok {
					continue
				}
				members = append(members, jid)
			}
		}

		return a.Store.CreateGroup(
			from, subject, creator, members)
	}
}
func New_Hook_GroupLeave(a *Acc) func(...any) error {
	// Decrypt and Replace node.Data
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		// MUST have attr `from`,`participant`
		from, ok1 := attrs[`from`] // gid
		type_, ok2 := attrs[`type`]

		if !ok1 || !ok2 {
			return nil
		}
		if type_ != `w:gp2` {
			return nil
		}
		chRemove, ok := n.FindChildByTag(`remove`)
		if !ok {
			return nil
		}
		my_jid, e := a.Store.GetMyJid()
		if e != nil {
			return e
		}
		for _, ch := range chRemove.Children {
			if ch.Tag == `participant` {
				m := ch.MapAttrs()
				if jid, ok := m[`jid`]; ok {
					if jid == my_jid { // self left group, clear group/members
						if e := a.Store.RemoveGroup(from); e != nil {
							return e
						}
					} else { // other leave group
						if e := a.Store.RemoveOneGroupMember(from, jid); e != nil {
							return e
						}

					}
				}
			}
		}

		return nil
	}
}
func New_Hook_GroupAdd(a *Acc) func(...any) error {
	return func(args ...any) error {
		n := args[0].(*xmpp.Node)

		attrs := n.MapAttrs()

		from, ok1 := attrs[`from`] // gid
		type_, ok2 := attrs[`type`]

		if !ok1 || !ok2 {
			return nil
		}
		if type_ != `w:gp2` {
			return nil
		}
		chAdd, ok := n.FindChildByTag(`add`)
		if !ok {
			return nil
		}
		for _, ch := range chAdd.Children {
			if ch.Tag == `participant` {
				m := ch.MapAttrs()
				if jid, ok := m[`jid`]; ok {
					if e := a.Store.AddGroupMember(from, jid); e != nil {
						return e
					}
				}
			}
		}

		return nil
	}
}

func (c Core) GroupAdmin(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	jids, e := j.Get("jids").TryStringArray()
	if e != nil {
		return NewErrRet(errors.New(`wrong param jids, no string array`))
	}
	type_, e := j.Get("type").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param 'type', not string`))
	}

	ch := []*xmpp.Node{}
	for _, jid := range jids {
		ch = append(ch, &xmpp.Node{
			Tag: `participant`,
			Attrs: []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			},
		})
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `admin`,
				Children: []*xmpp.Node{
					{
						Tag:      type_,
						Children: ch,
					},
				},
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}

func (c Core) GroupGetInvite(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}

	ch := &xmpp.Node{
		Tag: `invite`,
	}

	// Param
	if code, err := j.Get("code").TryString(); err == nil {
		ch.Attrs = []*xmpp.KeyValue{
			{Key: `code`, Value: code},
		}
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
func (c Core) GroupSetInvite(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}

	ch := &xmpp.Node{
		Tag: `invite`,
	}

	if code, err := j.Get("code").TryString(); err == nil {
		ch.Attrs = []*xmpp.KeyValue{
			{Key: `code`, Value: code},
		}
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}

func (c Core) GroupAnnouncement(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	gid, e := j.Get("gid").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param gid`))
	}
	type_, e := j.Get("type").TryString()
	if e != nil {
		return NewErrRet(errors.New(`wrong param 'type', not string`))
	}

	nr, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: gid},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:g2`},
		},
		Children: []*xmpp.Node{
			{
				Tag: type_,
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(nr.ToJson())
}
