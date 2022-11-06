package core

import (
	"ajson"
	"algo"
	"strings"
	"wa/xmpp"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

func (a *Acc) get_biz_profile(jid, v string) error {
	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:biz`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `business_profile`,
				Attrs: []*xmpp.KeyValue{
					{Key: `v`, Value: v},
				},
				Children: []*xmpp.Node{
					{
						Tag: `profile`,
						Attrs: []*xmpp.KeyValue{
							{Key: `jid`, Value: jid},
						},
					},
				},
			},
		},
	})
	return e
}

func (a *Acc) set_biz_profile(j *ajson.Json) error {
	chs := []*xmpp.Node{}

	if j.Exists(`BizAddress`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `address`,
			Data: []byte(j.Get(`BizAddress`).String()),
		})
	}
	if j.Exists(`BizLongitude`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `longitude`,
			Data: []byte(j.Get(`BizLongitude`).String()),
		})
	}
	if j.Exists(`BizLatitude`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `latitude`,
			Data: []byte(j.Get(`BizLatitude`).String()),
		})
	}

	if j.Exists(`BizDescription`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `description`,
			Data: []byte(j.Get(`BizDescription`).String()),
		})
	}
	if j.Exists(`BizWebsite`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `website`,
			Data: []byte(j.Get(`BizWebsite`).String()),
		})
	}
	if j.Exists(`BizEmail`) {
		chs = append(chs, &xmpp.Node{
			Tag:  `email`,
			Data: []byte(j.Get(`BizEmail`).String()),
		})
	}

	if j.Exists(`BizCategory`) {
		catgs := []*xmpp.Node{}

		if cat, e := j.Get(`BizCategory`).TryString(); e == nil {
			catgs = append(catgs, &xmpp.Node{
				Tag: `category`,
				Attrs: []*xmpp.KeyValue{
					{Key: `id`, Value: cat},
				},
			})
		} else {

			for _, cat := range j.Get(`BizCategory`).StringArray() {
				catgs = append(catgs, &xmpp.Node{
					Tag: `category`,
					Attrs: []*xmpp.KeyValue{
						{Key: `id`, Value: cat},
					},
				})
				break // only append the first 1
			}
		}

		chs = append(chs, &xmpp.Node{
			Tag:      `categories`,
			Children: catgs,
		})
	}

	_, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:biz`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `business_profile`,
				Attrs: []*xmpp.KeyValue{
					{Key: `v`, Value: `372`}, // 116:"biz_profile_options"
				},
				Children: chs,
			},
		},
	})
	return e
}
func (c Core) SetBizProfile(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}
	e = a.set_biz_profile(j)
	if e != nil {
		return NewErrRet(e)
	}
	return NewSucc()
}

func (a *Acc) get_profile_picture(j *ajson.Json) (*xmpp.Node, error) {
	// Param
	jid := j.Get(`jid`).String()

	ch := &xmpp.Node{
		Tag: `picture`,
	}

	// Param
	type_, e := j.Get(`type`).TryString()
	if e == nil {
		attrs := []*xmpp.KeyValue{}

		// Param
		query, e := j.Get(`query`).TryString()
		if e == nil {
			attrs = append(attrs, &xmpp.KeyValue{Key: `query`, Value: query})
		}
		// Param
		id, e := j.Get(`id`).TryString()
		if e == nil {
			attrs = append(attrs, &xmpp.KeyValue{Key: `id`, Value: id})
		}

		attrs = append(attrs, &xmpp.KeyValue{Key: `type`, Value: type_})
		ch.Attrs = attrs
	} else {

		// Param
		invite, e := j.Get(`invite`).TryString()
		if e == nil {
			attrs := []*xmpp.KeyValue{
				{Key: `invite`, Value: invite},
			}
			ch.Attrs = attrs
		}
	}

	return a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: jid, Type: 1},
			{Key: `type`, Value: `get`},
			{Key: `xmlns`, Value: `w:profile:picture`},
		},
		Children: []*xmpp.Node{
			ch,
		},
	})
}
func (c Core) SetProfilePicture(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	jid := j.Get(`jid`).String()
	img, e := algo.B64Dec(j.Get(`img`).String())
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail decode 'img'`))
	}

	n := &xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:profile:picture`},
		},
		Children: []*xmpp.Node{
			{
				Tag: `picture`,
				Attrs: []*xmpp.KeyValue{
					{Key: `type`, Value: `image`},
				},
				Data: img,
			},
		},
	}
	// 1 more attr `target` for group
	if strings.HasSuffix(jid, "@g.us") {
		n.Attrs = append(n.Attrs, &xmpp.KeyValue{
			Key: `target`, Value: jid,
		})
	}
	rn, e := a.Noise.WriteReadXmppNode(n)
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(rn.ToJson())
}

func (c Core) GetProfilePicture(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	rn, e := a.get_profile_picture(j)

	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(rn.ToJson())
}

func (c Core) SetMyProfile(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	mod := bson.M{}

	if avatar_b64, e := j.Get(`Avatar`).TryString(); e == nil {
		val, e := algo.B64Dec(avatar_b64)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `fail decode 'Avatar'`))
		}
		mod[`Avatar`] = val
	}

	if val, e := j.Get(`Nick`).TryString(); e == nil {
		mod[`Nick`] = val
	}
	if val, e := j.Get(`BizAddress`).TryString(); e == nil {
		mod[`BizAddress`] = val
	}
	if val, e := j.Get(`BizDescription`).TryString(); e == nil {
		mod[`BizDescription`] = val
	}
	if val, e := j.Get(`BizCategory`).TryString(); e == nil {
		mod[`BizCategory`] = val
	} else if val, e := j.Get(`BizCategory`).TryStringArray(); e == nil {
		mod[`BizCategory`] = val[0]
	}
	if len(mod) == 0 {
		return NewSucc()
	}
	e = a.Store.ModifyProfile(mod)
	if e != nil {
		return NewErrRet(e)
	}

	return NewSucc()
}

func (c Core) SetStatus(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	status, e := j.Get(`status`).TryString()
	if e != nil {
		return NewErrRet(errors.Wrap(e, `wrong param 'status'`))
	}
	rn, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `status`},
		},
		Children: []*xmpp.Node{
			{
				Tag:  `status`,
				Data: []byte(status),
			},
		},
	})
	if e != nil {
		return NewErrRet(e)
	}

	return NewJsonRet(rn.ToJson())
}
