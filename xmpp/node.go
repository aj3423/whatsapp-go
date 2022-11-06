package xmpp

import (
	"bytes"
	"reflect"
	"strings"

	"ahex"
	"ajson"

	"github.com/pkg/errors"
)

type KeyValue struct {
	Key   string
	Value string
	Type  int
}

func (kv *KeyValue) IsCompliant() bool {
	return kv.Type == 1 ||
		strings.Contains(kv.Value, "s.whatsapp.net") ||
		strings.Contains(kv.Value, "g.us")
}

type Node struct {
	Compressed bool

	Tag      string
	Attrs    []*KeyValue
	Data     []byte
	Children []*Node
}

const INDENT = `    `

func (n1 *Node) Compare(n2 *Node) bool {
	if n1.Tag != n2.Tag {
		return false
	}
	if !bytes.Equal(n1.Data, n2.Data) {
		return false
	}
	if !reflect.DeepEqual(n1.Attrs, n2.Attrs) {
		return false
	}
	if !reflect.DeepEqual(n1.Children, n2.Children) {
		return false
	}
	return true
}

func (n *Node) ToString() string {
	return n.ToJson().ToStringIndent()
}
func (n *Node) ToJson() *ajson.Json {
	ret := ajson.New()

	ret.Set("Tag", n.Tag)
	if len(n.Attrs) > 0 {
		attr := ajson.New()
		for _, a := range n.Attrs {
			attr.Set(a.Key, a.Value)
		}
		ret.Set(`Attrs`, attr)
	}
	if n.Data != nil {
		ret.Set("Data", ahex.Enc(n.Data))
	}
	if len(n.Children) > 0 {
		children := []interface{}{}

		for _, c := range n.Children {
			children = append(children, c.ToJson().Data())
		}
		ret.Set("Children", children)
	}
	return ret
}
func ParseJson(j *ajson.Json) (*Node, error) {
	n := &Node{}

	// Tag
	tag, e := j.Get(`Tag`).TryString()
	if e != nil {
		return nil, errors.New(`fail parse Tag`)
	}
	n.Tag = tag

	// Attrs
	if j.Exists(`Attrs`) {
		attrs, ok := j.TryGet(`Attrs`)
		if !ok {
			return nil, errors.New(`invalid "Attrs"`)
		}
		m, e := attrs.TryMap()
		if e != nil {
			return nil, errors.New(`fail parse "Attrs" to map`)
		}
		for k, v := range m {
			vs, ok := v.(string)
			if !ok {
				return nil, errors.New(`fail cast Attr "` + k + `" to string`)
			}
			n.Attrs = append(n.Attrs, &KeyValue{Key: k, Value: vs})
		}
	}

	// Data
	if j.Exists(`Data`) {
		data, e := j.Get(`Data`).TryString()
		if e != nil {
			return nil, e
		}
		n.Data = ahex.Dec(data)
	}

	// Children
	if j.Exists(`Children`) {
		chs, e := j.Get(`Children`).TryJsonArray()
		if e != nil {
			return nil, e
		}
		for _, ch := range chs {
			nn, e := ParseJson(ch)
			if e != nil {
				return nil, e
			}
			n.Children = append(n.Children, nn)
		}
	}

	return n, nil
}

func (n *Node) GetAttr(attr string) (string, bool) {
	for _, a := range n.Attrs {
		if a.Key == attr {
			return a.Value, true
		}
	}
	return ``, false
}
func (n *Node) SetAttr(attr, value string) {
	for _, a := range n.Attrs {
		if a.Key == attr {
			a.Value = value
			return
		}
	}
	// if not found, add one
	n.Attrs = append(n.Attrs, &KeyValue{
		Key: attr, Value: value,
	})
}
func (n *Node) MapAttrs() map[string]string {
	m := map[string]string{}
	for _, attr := range n.Attrs {
		m[attr.Key] = attr.Value
	}
	return m
}
func (n *Node) GetAttrs(keys []string) ([]string, error) {
	m := n.MapAttrs()

	ret := []string{}
	for _, key := range keys {
		val, ok := m[key]
		if !ok {
			return nil, errors.New(`missing ` + key)
		}
		ret = append(ret, val)
	}
	return ret, nil
}

func (n *Node) FindChild(fn func(*Node) bool) (*Node, bool) {
	for _, ch := range n.Children {
		if fn(ch) {
			return ch, true
		}
	}
	return nil, false
}
func (n *Node) FindChildByTag(tag string) (*Node, bool) {
	return n.FindChild(func(ch *Node) bool {
		return ch.Tag == tag
	})
}
func (n *Node) FindChildWithAttr(key, val string) (*Node, bool) {
	return n.FindChild(func(ch *Node) bool {
		attr, ok := ch.GetAttr(key)
		return ok && attr == val
	})
}
