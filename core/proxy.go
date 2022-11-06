package core

import (
	"encoding/json"

	"ajson"
)

func (c Core) SetProxy(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	if addr, e := j.Get(`Addr`).TryString(); e == nil {
		if e := a.Store.SetProxy(addr); e != nil {
			return NewErrRet(e)
		}
		a.Noise.Socket.Proxy = addr
		a.Log.Info("SetProxy: " + addr)
	}

	if dns, e := j.Get(`Dns`).TryMap(); e == nil {
		// copy to map
		m := map[string]string{}
		for k, v := range dns {
			m[k] = v.(string)
		}
		if e := a.Store.SetDns(m); e != nil {
			return NewErrRet(e)
		}
		a.Noise.Socket.Dns = m
		{ // log
			bs, _ := json.Marshal(m)
			a.Log.Info("SetDns: " + string(bs))
		}
	}

	return NewRet(0)
}
