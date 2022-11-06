package core

/*
func (c Core) _Req(j *ajson.Json, wait_for_resp bool) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	n, e := xmpp.ParseJson(j)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `fail parse json`))
	}

	if wait_for_resp {
		// 1.
		//   if wait_for_resp but no `id` provided
		//   auto generate one
		id, ok := n.GetAttr(`id`)
		if !ok || id == `auto` || id == `` {
			n.SetAttr(`id`, a.Noise.NextIqId_2())
		}

		// 2. send
		nr, e := a.Noise.WriteReadXmppNode(n)
		if e != nil {
			return NewErrRet(e)
		}
		return NewJsonRet(nr.ToJson())
	} else {
		e := a.Noise.WriteXmppNode(n)
		if e != nil {
			return NewErrRet(e)
		}
		return NewSucc()
	}

}
func (c Core) Req(j *ajson.Json) *ajson.Json {
	return c._Req(j, false)
}
func (c Core) ReqResp(j *ajson.Json) *ajson.Json {
	return c._Req(j, true)
}
*/
