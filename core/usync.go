package core

import (
	"strconv"
	"time"

	"ajson"
	"arand"
	"wa/xmpp"

	"github.com/pkg/errors"
)

/*
	FirstUpload 	Refresh 	Background

context:	`delta`			`full`		`delta`
*/
func (a *Acc) do_usync(
	mode string, // `full`/`delta`
	context string, // interactive/background
	sid string,
	children []*xmpp.Node,
) (*xmpp.Node, error) {
	n := &xmpp.Node{
		Compressed: true,

		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `usync`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `usync`,
				Attrs: []*xmpp.KeyValue{
					{Key: `context`, Value: context},
					{Key: `index`, Value: `0`},
					{Key: `last`, Value: `true`},
					{Key: `mode`, Value: mode},
					{Key: `sid`, Value: `sync_sid_` + sid + `_` + arand.Uuid4()},
				},
				Children: children,
			},
		},
	}
	nr, e := a.Noise.WriteReadXmppNode(n)
	return nr, e
}

/*
used when list_contact

	FirstUpload 	Refresh 	Background

context:	`delta`			`full`		`delta`
*/
func (a *Acc) usync_contact(
	mode string, // `full`/`delta`
	context string, // interactive/background
	sid string,
	users []*ajson.Json,
) (*xmpp.Node, error) {

	list := []*xmpp.Node{}
	for _, u := range users {
		contact, e := u.Get(`contact`).TryString()
		if e != nil {
			return nil, errors.New(`attr contact not string`)
		}
		n := &xmpp.Node{
			Tag: `user`,
			Children: []*xmpp.Node{
				{
					Tag:  `contact`,
					Data: []byte(contact),
				},
			},
		}
		if jid, e := u.Get(`jid`).TryString(); e == nil {
			n.Attrs = []*xmpp.KeyValue{
				{Key: `jid`, Value: jid},
			}
		}
		list = append(list, n)
	}

	return a.do_usync(mode, context, sid, []*xmpp.Node{
		{
			Tag: `query`,
			Children: []*xmpp.Node{
				contact_node(),
				status_node(),
				business_profile_node(),
				devices_version_node(),
				disappearing_mode_node(),
			},
		},
		{
			Tag:      `list`,
			Children: list,
		},
	})
}

// TODO, children contain all group member?
func (a *Acc) usync_device(
	mode string, // `query`
	context string, // notification
	sid string, // devices
) (*xmpp.Node, error) {

	dev, e := a.Store.GetDev()
	if e != nil {
		return nil, e
	}

	my_jid := dev.Cc + dev.Phone + `@s.whatsapp.net`
	ch := []*xmpp.Node{
		{
			Tag: `query`,
			Children: []*xmpp.Node{
				devices_version_node(),
			},
		},
		{
			Tag: `list`,
			Children: []*xmpp.Node{
				{
					Tag: `user`,
					Attrs: []*xmpp.KeyValue{
						{Key: `jid`, Value: my_jid},
					},
					Children: []*xmpp.Node{
						devices_hash_node(my_jid),
					},
				},
			},
		},
	}
	return a.do_usync(mode, context, sid, ch)
}

// not necessary
// usync after receiving msg from web client( from xxx.0:1@s.whatsapp.net )
// or peer device `update` notification
func (a *Acc) usync_multi_device(
	jid string, // 111.0:2@s.whatsapp.net
) (*xmpp.Node, error) {
	ch := []*xmpp.Node{
		{
			Tag: `query`,
			Children: []*xmpp.Node{
				devices_version_node(),
			},
		},
		{
			Tag: `list`,
			Children: []*xmpp.Node{
				{
					Tag: `user`,
					Attrs: []*xmpp.KeyValue{
						// only use phone jid
						{Key: `jid`, Value: clear_jid_device(jid)},
					},
					Children: []*xmpp.Node{
						multi_devices_node(jid, time.Now().Add(-5*time.Minute)),
					},
				},
			},
		},
	}
	return a.do_usync("query", "interactive", "multi_protocols", ch)
}

// handle above result
func (a *Acc) handle_usync_multi_device_result(
	nr *xmpp.Node, recid uint64,
) error {

	n_usync, ok := nr.FindChildByTag(`usync`)
	if !ok {
		return nil
	}
	n_list, ok := n_usync.FindChildByTag(`list`)
	if !ok {
		return nil
	}
	n_user, ok := n_list.FindChildByTag(`user`)
	if !ok {
		return nil
	}
	n_devices, ok := n_user.FindChildByTag(`devices`)
	if !ok {
		return nil
	}
	n_device_list, ok := n_devices.FindChildByTag(`device_list`)
	if !ok {
		return nil
	}

	// remove all devids first
	if e := a.Store.DelAllMultiDevice(recid); e != nil {
		return e
	}
	// add each id
	for _, dev := range n_device_list.Children {
		if dev.Tag == `device` {
			id, ok := dev.GetAttr(`id`)
			if !ok || id == `0` { // skip id 0 as phone
				continue
			}
			// add to db if not exists
			devid, e := strconv.Atoi(id)
			if e != nil {
				return e
			}
			a.Store.AddMultiDevice(recid, uint32(devid))
			a.Store.SetMultiDeviceLastSync(recid, uint32(devid))
		}
	}
	return nil
}
func (c Core) ListContact(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	mode := j.Get(`mode`).String()
	context := j.Get(`context`).String()
	sid := j.Get(`sid`).String()
	list := j.Get(`list`).JsonArray()
	nr, e := a.usync_contact(mode, context, sid, list)
	if e != nil {
		return NewErrRet(e)
	}
	return NewJsonRet(nr.ToJson())
}

// used scenario:
// 1. recv stranger msg, add him as contact
func (c Core) UsyncProfile(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	mode := j.Get(`mode`).String()
	context := j.Get(`context`).String()
	sid := j.Get(`sid`).String()

	jid := j.Get(`jid`).String()

	nr, e := a.do_usync(
		mode,
		context,
		sid,
		[]*xmpp.Node{
			{
				Tag: `query`,
				Children: []*xmpp.Node{
					status_node(),
					business_profile_node(),
					sidelist_node(),
					devices_version_node(),
					disappearing_mode_node(),
				},
			},
			{
				Tag: `side_list`,
				Children: []*xmpp.Node{
					{
						Tag: `user`,
						Attrs: []*xmpp.KeyValue{
							{Key: `jid`, Value: jid},
						},
						Children: []*xmpp.Node{
							devices_hash_node(jid),
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

func (c Core) QueryChatLink(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	mode := "query"
	context := "interactive"
	sid := "query"

	phone := j.Get(`phone`).String()

	nr, e := a.do_usync(
		mode,
		context,
		sid,
		[]*xmpp.Node{
			{
				Tag: `query`,
				Children: []*xmpp.Node{
					contact_node(),
					status_node(),
					business_profile_node(),
					picture_preview_node(),
					disappearing_mode_node(),
				},
			},
			{
				Tag: `list`,
				Children: []*xmpp.Node{
					{
						Tag: `user`,
						Children: []*xmpp.Node{
							{
								Tag:  `contact`,
								Data: []byte(phone),
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

func contact_node() *xmpp.Node {
	return &xmpp.Node{Tag: `contact`}
}
func status_node() *xmpp.Node {
	return &xmpp.Node{Tag: `status`}
}
func disappearing_mode_node() *xmpp.Node {
	return &xmpp.Node{Tag: `disappearing_mode`}
}
func picture_preview_node() *xmpp.Node {
	return &xmpp.Node{
		Tag: `picture`,
		Attrs: []*xmpp.KeyValue{
			{Key: `type`, Value: `preview`},
		},
	}
}
func sidelist_node() *xmpp.Node {
	return &xmpp.Node{Tag: `sidelist`}
}
func business_profile_node() *xmpp.Node {
	return &xmpp.Node{
		Tag: `business`,
		Children: []*xmpp.Node{
			{
				Tag: `verified_name`,
			},
			{
				Tag: `profile`,
				Attrs: []*xmpp.KeyValue{
					{Key: `v`, Value: `116`},
				},
			},
		},
	}
}
func devices_version_node() *xmpp.Node {
	return &xmpp.Node{
		Tag: `devices`,
		Attrs: []*xmpp.KeyValue{
			{Key: `version`, Value: `2`},
		},
	}
}
func devices_hash_node(jid string) *xmpp.Node {
	return &xmpp.Node{
		Tag: `devices`,
		Attrs: []*xmpp.KeyValue{
			{Key: `device_hash`, Value: phash(jid, nil)},
		},
	}
}
func multi_devices_node(jid string, last_sync time.Time) *xmpp.Node {

	return &xmpp.Node{
		Tag: `devices`,
		Attrs: []*xmpp.KeyValue{
			{Key: `device_hash`, Value: phash(jid, nil)},
			// the expected_ts should use the 'ts' in the proto from device change notification
			// which contains:   "Tag": "device-identity" proto
			// just ignore this attribute, many requests captured without this field
			// {Key: `expected_ts`, Value: strconv.Itoa(int(time.Now().Unix()))},
			//{Key: `expected_ts`, Value: strconv.Itoa(int(1650715908))},
			{Key: `ts`, Value: strconv.Itoa(int(last_sync.Unix()))},
		},
	}
}
