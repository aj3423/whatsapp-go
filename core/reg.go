package core

import (
	"ajson"
	"algo"
	"arand"
	"wa/crypto"
	"wa/def"
	"wa/net"

	fhttp "github.com/useflyent/fhttp"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var reg_host = "v.whatsapp.net"

func reg_http_header(host, ua string) fhttp.Header {
	m := fhttp.Header{
		"User-Agent":      {ua},
		"WaMsysRequest":   {"1"},
		"request_token":   {arand.Uuid4()},
		"Host":            {host},
		"Connection":      {"Keep-Alive"},
		"Accept-Encoding": {"gzip"},

		fhttp.HeaderOrderKey: {
			"user-agent",
			"wamsysrequest",
			"request_token",
			"host",
			"connection",
			"accept-encoding",
		},
	}
	return m
}

func (c Core) Exist(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	priv, pub := crypto.NewECKeyPair()

	agreed := crypto.Curve25519Agree(
		priv, def.SVR_PUB,
	)

	if j.Exists(`recovery_token`) {
		rc_b64, e := j.Get(`recovery_token`).TryString()
		if e != nil {
			return NewErrRet(errors.New(`'recovery_token' wrong format`))
		}
		rc, e := algo.B64Dec(rc_b64)
		if e != nil {
			return NewErrRet(errors.New(`'recovery_token' not base64`))
		}
		if e := a.Store.ModifyDev(bson.M{
			`RecoveryToken`: rc,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `set 'recovery_token' fail`))
		}
	}
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}

	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return NewErrRet(e)
	}
	spk_sig := spk.Signature()
	spk_key := spk.KeyPair().PublicKey().PublicKey()
	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return NewErrRet(e)
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	p := map[string]string{}
	if grnt, e := j.Get(`read_phone_permission_granted`).TryString(); e == nil {
		p["read_phone_permission_granted"] = grnt
	} else {
		p["read_phone_permission_granted"] = `0`
	}
	p["lc"] = dev.Locale                           // CN
	p["offline_ab"] = j.Get("offline_ab").String() // {exposure...

	if j.Exists(`phone`) { // for checking block
		p["in"] = j.Get(`phone`).String()
	} else {
		p["in"] = dev.Phone
	}
	p["backup_token"] = string(dev.BackupToken)
	p["lg"] = dev.Language //zh
	p["e_regid"] = algo.B64RawUrlEnc(crypto.U322BE(dev.RegId))
	p["mistyped"] = j.Get("mistyped").String()
	p["id"] = string(dev.RecoveryToken)
	p["authkey"] = algo.B64RawUrlEnc(cfg.StaticPub)
	p["e_skey_sig"] = algo.B64RawUrlEnc(spk_sig[:])
	p["token"] = algo.B64Enc(crypto.CalcToken(dev.Phone, dev.IsBusiness)) // this is std, others are all B64RawUrlEnc
	p["expid"] = algo.B64RawUrlEnc(dev.ExpId)
	p["e_ident"] = algo.B64RawUrlEnc(iden_pub[:])
	p["rc"] = j.Get("rc").String()
	p["simnum"] = j.Get("simnum").String()
	p["sim_state"] = j.Get("sim_state").String()
	p["client_metrics"] = j.Get("client_metrics").String()
	if j.Exists(`cc`) { // for checking block
		p["cc"] = j.Get(`cc`).String()
	} else {
		p["cc"] = dev.Cc //86
	}
	p["e_skey_id"] = algo.B64RawUrlEnc([]byte{0, 0, 0})
	p["fdid"] = dev.Fdid // uuid4
	p["e_skey_val"] = algo.B64RawUrlEnc(spk_key[:])
	p["hasinrc"] = j.Get("hasinrc").String()
	p["network_radio_type"] = j.Get("network_radio_type").String()
	p["network_operator_name"] = dev.NetworkOperatorName
	p["sim_operator_name"] = dev.SimOperatorName
	p["e_keytype"] = algo.B64RawUrlEnc([]byte{5}) // 05
	p["pid"] = j.Get("pid").String()

	url, e := crypto.BuildUrl(p)
	if e != nil {
		return NewErrRet(e)
	}
	a.Log.Info(`Exist() url: %s`, url)

	ENC, e := algo.AesGcmEnc([]byte(url), agreed, def.IV, def.AAD)
	if e != nil {
		return NewErrRet(e)
	}
	ENC = append(pub, ENC...)

	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return NewErrRet(e)
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	status, body, e := net.HttpGet(
		reg_host,
		"/v2/exist?ENC="+algo.B64RawUrlEnc(ENC),
		proxy, dns, dev.Ja3Config,
		reg_http_header(reg_host, ua),
	)
	if e != nil {
		a.Log.Error(`Exist() error: %s`, e.Error())
		return NewErrRet(e)
	}
	a.Log.Success(`Exist() ret: %s`, string(body))

	rj, e := ajson.ParseByte(body)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `return body: `+string(body)))
	}
	r := NewJsonRet(rj)
	r.Set(`Status`, status)
	return r
}

func (c Core) ClientLog(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	priv, pub := crypto.NewECKeyPair()

	agreed := crypto.Curve25519Agree(
		priv, def.SVR_PUB,
	)
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}
	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return NewErrRet(e)
	}
	spk_sig := spk.Signature()
	spk_key := spk.KeyPair().PublicKey().PublicKey()

	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return NewErrRet(e)
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	p := map[string]string{}
	p["lc"] = dev.Locale // CN
	p["in"] = dev.Phone  // 15322223333
	p["backup_token"] = string(dev.BackupToken)
	p["lg"] = dev.Language // zh
	p["id"] = string(dev.RecoveryToken)
	p["e_regid"] = algo.B64RawUrlEnc(crypto.U322BE(dev.RegId))
	p["authkey"] = algo.B64RawUrlEnc(cfg.StaticPub)
	p["e_skey_sig"] = algo.B64RawUrlEnc(spk_sig[:])
	p["action_taken"] = j.Get(`action_taken`).String()
	p["expid"] = algo.B64RawUrlEnc(dev.ExpId)
	p["e_ident"] = algo.B64RawUrlEnc(iden_pub[:])
	p["previous_screen"] = j.Get(`previous_screen`).String()
	p["cc"] = dev.Cc // 86
	p["e_skey_id"] = algo.B64RawUrlEnc([]byte{0, 0, 0})
	p["fdid"] = dev.Fdid // uuid4
	p["funnel_id"] = arand.Uuid4()
	p["e_skey_val"] = algo.B64RawUrlEnc(spk_key[:])
	p["e_keytype"] = algo.B64RawUrlEnc([]byte{5})      // 05
	p["current_screen"] = algo.B64RawUrlEnc([]byte{5}) // 05

	url, e := crypto.BuildUrl(p)
	if e != nil {
		return NewErrRet(e)
	}

	a.Log.Info(`ClientLog() url: %s`, url)

	ENC, e := algo.AesGcmEnc([]byte(url), agreed, def.IV, def.AAD)
	if e != nil {
		return NewErrRet(e)
	}
	ENC = append(pub, ENC...)

	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return NewErrRet(e)
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	status, body, e := net.HttpGet(
		reg_host,
		"/v2/client_log?ENC="+algo.B64RawUrlEnc(ENC),
		proxy, dns, dev.Ja3Config,
		reg_http_header(reg_host, ua),
	)
	if e != nil {
		a.Log.Error(`ClientLog() error: %s`, e.Error())
		return NewErrRet(e)
	}
	a.Log.Success(`ClientLog() ret: %s`, string(body))
	rj, e := ajson.ParseByte(body)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `return body: `+string(body)))
	}
	r := NewJsonRet(rj)
	r.Set(`Status`, status)
	return r
}

func (c Core) Code(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	priv, pub := crypto.NewECKeyPair()

	agreed := crypto.Curve25519Agree(
		priv, def.SVR_PUB,
	)
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}
	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return NewErrRet(e)
	}
	spk_sig := spk.Signature()
	spk_key := spk.KeyPair().PublicKey().PublicKey()

	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return NewErrRet(e)
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	p := map[string]string{}
	p[`reason`] = ``
	p[`method`] = j.Get(`method`).String()
	p["lc"] = dev.Locale // CN
	p["in"] = dev.Phone  // 15322223333
	p["backup_token"] = string(dev.BackupToken)
	p["lg"] = dev.Language // zh
	p["e_regid"] = algo.B64RawUrlEnc(crypto.U322BE(dev.RegId))
	p["mistyped"] = j.Get("mistyped").String()
	p["id"] = string(dev.RecoveryToken)
	p["authkey"] = algo.B64RawUrlEnc(cfg.StaticPub)
	p["e_skey_sig"] = algo.B64RawUrlEnc(spk_sig[:])
	p["hasav"] = j.Get(`hasav`).String()
	p["token"] = algo.B64Enc(crypto.CalcToken(dev.Phone, dev.IsBusiness)) // this is std, others are all B64RawUrlEnc
	p["expid"] = algo.B64RawUrlEnc(dev.ExpId)
	p["e_ident"] = algo.B64RawUrlEnc(iden_pub[:])
	p["rc"] = j.Get("rc").String()
	p["sim_mcc"] = dev.SimMcc
	p["simnum"] = j.Get("simnum").String()
	p["client_metrics"] = j.Get("client_metrics").String()
	p["cc"] = dev.Cc // 86
	p["e_skey_id"] = algo.B64RawUrlEnc([]byte{0, 0, 0})
	p["mnc"] = dev.Mnc
	p["sim_mnc"] = dev.SimMnc
	p["fdid"] = dev.Fdid // uuid4
	p["e_skey_val"] = algo.B64RawUrlEnc(spk_key[:])
	p["hasinrc"] = j.Get("hasinrc").String()
	p["network_radio_type"] = j.Get("network_radio_type").String()
	p["mcc"] = dev.Mcc
	p["e_keytype"] = algo.B64RawUrlEnc([]byte{5}) // 05
	p["pid"] = j.Get("pid").String()

	url, e := crypto.BuildUrl(p)
	if e != nil {
		return NewErrRet(e)
	}

	a.Log.Info(`Code() url: %s`, url)

	ENC, e := algo.AesGcmEnc([]byte(url), agreed, def.IV, def.AAD)
	if e != nil {
		return NewErrRet(e)
	}
	ENC = append(pub, ENC...)

	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return NewErrRet(e)
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	status, body, e := net.HttpGet(
		reg_host,
		"/v2/code?ENC="+algo.B64RawUrlEnc(ENC),
		proxy, dns, dev.Ja3Config,
		reg_http_header(reg_host, ua),
	)
	if e != nil {
		a.Log.Error(`Code() error: %s`, e.Error())
		return NewErrRet(e)
	}
	a.Log.Success(`Code() ret: %s`, string(body))
	rj, e := ajson.ParseByte(body)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `return body: `+string(body)))
	}
	r := NewJsonRet(rj)
	r.Set(`Status`, status)
	return r
}

func (c Core) Register(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	priv, pub := crypto.NewECKeyPair()

	agreed := crypto.Curve25519Agree(
		priv, def.SVR_PUB,
	)
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}

	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return NewErrRet(e)
	}
	spk_sig := spk.Signature()
	spk_key := spk.KeyPair().PublicKey().PublicKey()

	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return NewErrRet(e)
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	p := map[string]string{}
	p["lc"] = dev.Locale // CN
	p["in"] = dev.Phone  // 15322223333
	p["backup_token"] = string(dev.BackupToken)
	p["lg"] = dev.Language // zh
	p["e_regid"] = algo.B64RawUrlEnc(crypto.U322BE(dev.RegId))
	p["mistyped"] = j.Get("mistyped").String()
	p["id"] = string(dev.RecoveryToken)
	p["authkey"] = algo.B64RawUrlEnc(cfg.StaticPub)
	p["e_skey_sig"] = algo.B64RawUrlEnc(spk_sig[:])
	p["expid"] = algo.B64RawUrlEnc(dev.ExpId)
	p["e_ident"] = algo.B64RawUrlEnc(iden_pub[:])
	p["rc"] = j.Get("rc").String()
	p["sim_mcc"] = dev.SimMcc
	p["simnum"] = j.Get("simnum").String()
	p["entered"] = j.Get("entered").String()
	p["client_metrics"] = j.Get("client_metrics").String()
	p["cc"] = dev.Cc // 86
	p["e_skey_id"] = algo.B64RawUrlEnc([]byte{0, 0, 0})
	p["mnc"] = dev.Mnc
	p["sim_mnc"] = dev.SimMnc
	p["fdid"] = dev.Fdid // uuid4
	p["e_skey_val"] = algo.B64RawUrlEnc(spk_key[:])
	p["network_radio_type"] = j.Get("network_radio_type").String()
	p["hasinrc"] = j.Get("hasinrc").String()
	p["network_operator_name"] = dev.NetworkOperatorName
	p["sim_operator_name"] = dev.SimOperatorName
	p["mcc"] = dev.Mcc
	p["e_keytype"] = algo.B64RawUrlEnc([]byte{5}) // 05
	p["pid"] = j.Get("pid").String()

	if dev.IsBusiness { // vname
		vname, e := a.generate_vname()
		if e != nil {
			return NewErrRet(e)
		}
		bs, _ := proto.Marshal(vname)
		p["vname"] = algo.B64RawUrlEnc(bs)

		// save for Security
		if e := a.Store.ModifyConfig(bson.M{
			`VNameCert`: bs,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `fail save tmp cert(vname)`))
		}
	}
	p["code"] = j.Get("code").String()

	url, e := crypto.BuildUrl(p)
	if e != nil {
		return NewErrRet(e)
	}

	a.Log.Info(`Register() url: %s`, url)

	ENC, e := algo.AesGcmEnc([]byte(url), agreed, def.IV, def.AAD)
	if e != nil {
		return NewErrRet(e)
	}
	ENC = append(pub, ENC...)

	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return NewErrRet(e)
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	status, body, e := net.HttpGet(
		reg_host,
		"/v2/register?ENC="+algo.B64RawUrlEnc(ENC),
		proxy, dns, dev.Ja3Config,
		reg_http_header(reg_host, ua),
	)
	if e != nil {
		a.Log.Error(`Register() error: %s`, e.Error())
		return NewErrRet(e)
	}
	a.Log.Success(`Register() ret: %s`, string(body))
	rj, e := ajson.ParseByte(body)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `return body: `+string(body)))
	}

	has_pin := false
	if rj.Get(`reason`).String() == `security_code` {
		has_pin = true
	}
	/* ver > 2.21.21.19 has no routing_info
	// routing_info
	if !has_pin {
		routingInfo, e := rj.Get(`edge_routing_info`).TryString()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `no RoutingInfo returned: `+string(body)))
		}
		ri, e := algo.B64Dec(routingInfo)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `RoutingInfo not b64: `+string(body)))
		}
		if e := a.Store.ModifyConfig(bson.M{
			`RoutingInfo`: ri,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `fail SaveRoutingInfo`))
		}
	}
	*/
	if dev.IsBusiness && !has_pin { // cert
		vname, e := rj.Get(`cert`).TryString()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `no cert returned: `+string(body)))
		}
		vn, e := algo.B64Dec(vname)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `cert not b64: `+string(body)))
		}
		if e := a.Store.ModifyConfig(bson.M{
			`VNameCert`: vn,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `fail cert(vname)`))
		}
	}

	// reg success if goes here

	r := NewJsonRet(rj)
	r.Set(`Status`, status)

	return r
}

func (c Core) Security(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	priv, pub := crypto.NewECKeyPair()

	agreed := crypto.Curve25519Agree(
		priv, def.SVR_PUB,
	)
	dev, e := a.Store.GetDev()
	if e != nil {
		return NewErrRet(e)
	}
	cfg, e := a.Store.GetConfig()
	if e != nil {
		return NewErrRet(e)
	}

	spk, e := a.Store.LoadSignedPreKey(0)
	if e != nil {
		return NewErrRet(e)
	}
	spk_sig := spk.Signature()
	spk_key := spk.KeyPair().PublicKey().PublicKey()

	iden, e := a.Store.GetIdentityKeyPair()
	if e != nil {
		return NewErrRet(e)
	}
	iden_pub := iden.PublicKey().PublicKey().PublicKey()

	p := map[string]string{}
	p["lc"] = dev.Locale // CN
	p["in"] = dev.Phone  // 15322223333
	p["backup_token"] = string(dev.BackupToken)
	p["lg"] = dev.Language // zh
	p["id"] = string(dev.RecoveryToken)
	p["e_regid"] = algo.B64RawUrlEnc(crypto.U322BE(dev.RegId))
	p["authkey"] = algo.B64RawUrlEnc(cfg.StaticPub)
	p["e_skey_sig"] = algo.B64RawUrlEnc(spk_sig[:])
	p["expid"] = algo.B64RawUrlEnc(dev.ExpId)
	p["e_ident"] = algo.B64RawUrlEnc(iden_pub[:])
	p["rc"] = j.Get("rc").String()
	p["simnum"] = j.Get("simnum").String()
	p["client_metrics"] = j.Get("client_metrics").String()
	p["cc"] = dev.Cc // 86
	p["e_skey_id"] = algo.B64RawUrlEnc([]byte{0, 0, 0})
	p["fdid"] = dev.Fdid // uuid4
	p["e_skey_val"] = algo.B64RawUrlEnc(spk_key[:])
	p["network_radio_type"] = j.Get("network_radio_type").String()
	p["hasinrc"] = j.Get("hasinrc").String()
	p["e_keytype"] = algo.B64RawUrlEnc([]byte{5}) // 05
	p["pid"] = j.Get("pid").String()

	if dev.IsBusiness { // vname
		// saved when Register
		p["vname"] = algo.B64RawUrlEnc(cfg.VNameCert)
	}
	p["code"] = j.Get("code").String()

	url, e := crypto.BuildUrl(p)
	if e != nil {
		return NewErrRet(e)
	}

	a.Log.Info(`Security() url: %s`, url)

	ENC, e := algo.AesGcmEnc([]byte(url), agreed, def.IV, def.AAD)
	if e != nil {
		return NewErrRet(e)
	}
	ENC = append(pub, ENC...)

	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return NewErrRet(e)
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	status, body, e := net.HttpGet(
		reg_host,
		"/v2/security?ENC="+algo.B64RawUrlEnc(ENC),
		proxy, dns, dev.Ja3Config,
		reg_http_header(reg_host, ua),
	)
	if e != nil {
		a.Log.Error(`Security() error: %s`, e.Error())

		return NewErrRet(e)
	}
	a.Log.Success(`Security() ret: %s`, string(body))

	rj, e := ajson.ParseByte(body)
	if e != nil {
		return NewErrRet(errors.Wrap(e, `return body: `+string(body)))
	}

	// routing_info
	routingInfo, e := rj.Get(`edge_routing_info`).TryString()
	if e == nil { // sometimes has no routing_info
		ri, e := algo.B64Dec(routingInfo)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `RoutingInfo not b64: `+string(body)))
		}
		if e := a.Store.ModifyConfig(bson.M{
			`RoutingInfo`: ri,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `fail SaveRoutingInfo`))
		}
	}
	if dev.IsBusiness { // cert
		vname, e := rj.Get(`cert`).TryString()
		if e != nil {
			return NewErrRet(errors.Wrap(e, `no cert returned: `+string(body)))
		}
		vn, e := algo.B64Dec(vname)
		if e != nil {
			return NewErrRet(errors.Wrap(e, `cert not b64: `+string(body)))
		}
		if e := a.Store.ModifyConfig(bson.M{
			`VNameCert`: vn,
		}); e != nil {
			return NewErrRet(errors.Wrap(e, `fail cert(vname)`))
		}
	}

	r := NewJsonRet(rj)
	r.Set(`Status`, status)
	return r
}
