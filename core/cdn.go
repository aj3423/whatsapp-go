package core

import (
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"time"

	"ajson"
	"algo"
	"arand"
	"wa/xmpp"

	"wa/def"
	"wa/net"
	"wa/pb"

	"github.com/pkg/errors"
	fhttp "github.com/useflyent/fhttp"
	"go.mongodb.org/mongo-driver/bson"
)

func nc_cat(content_type pb.Media_Type, hash []byte, max_buckets int) string {
	// sticker uses fixed `1`
	if content_type == pb.Media_Sticker {
		return `1`
	}

	x := big.Int{}
	x.SetBytes(hash)

	y := big.NewInt(int64(max_buckets))

	z := big.Int{}
	z.Mod(&x, y)

	return strconv.Itoa(int(z.Int64()) + 100)
}
func cdn_http_header(host, ua string) fhttp.Header {
	m := fhttp.Header{
		"User-Agent":      {ua},
		"Host":            {host},
		"Accept-Encoding": {"identity"},
		"Connection":      {"Keep-Alive"},

		fhttp.HeaderOrderKey: {
			"user-agent",
			"host",
			"accept-encoding",
			"connection",
		},
	}
	return m
}

func (a *Acc) cdn_download(
	media Media,
) ([]byte, error) {
	dev, e := a.Store.GetDev()
	if e != nil {
		return nil, e
	}
	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return nil, e
	}
	cdn, e := a.Store.GetCdn()
	if e != nil {
		return nil, e
	}

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)

	direct_ip := `1`

	// 1. use cdn host
	host := media.CdnHost(cdn)
	// 2. if cdn host == 0, use url of proto.messageUrl
	if len(host) == 0 {

		direct_ip = `0`

		url_, err := url.Parse(media.MsgUrl())
		if err != nil {
			return nil, err
		}
		host = url_.Hostname()
	}

	hash := media.EncFileHash()
	url_ := fmt.Sprintf(
		"%s&direct_ip=%s&auth=%s&hash=%s&_nc_cat=%s&mode=auto",

		media.DirectPath(),
		direct_ip,
		cdn.Auth,
		algo.UrlEnc(algo.B64UrlEnc(hash)),
		nc_cat(media.Type(), hash, cdn.MaxBuckets),
	)

	a.Log.Debug("host: " + `https://` + host)
	a.Log.Debug("url: " + url_)

	_, body, e := net.HttpGet(
		host,
		url_,
		proxy, dns, dev.Ja3Config,
		cdn_http_header(host, ua),
	)
	if e != nil {
		return nil, e
	}

	return body, nil
}

func (a *Acc) cdn_upload(
	media Media,
	f_enc []byte,
) (string, string, error) {
	dev, e := a.Store.GetDev()
	if e != nil {
		return ``, ``, e
	}
	proxy, dns, e := a.Store.GetProxy()
	if e != nil {
		return ``, ``, e
	}
	cdn, e := a.Store.GetCdn()
	if e != nil {
		return ``, ``, e
	}

	direct_ip := `1` // 1: use cdn, 0: use messageUrl

	// 1. use cdn host
	host := media.CdnHost(cdn)

	// 2. if cdn host == 0, use url of proto.messageUrl
	if len(host) == 0 {
		direct_ip = `0`

		url_, err := url.Parse(media.MsgUrl())
		if err != nil {
			return ``, ``, err
		}
		host = url_.Hostname()

		if len(host) == 0 {
			host = "mmg.whatsapp.net"
		}
	}
	encFileHash := algo.Sha256(f_enc)

	token := algo.Sha256(
		append(cdn.UploadTokenRandomBytes, encFileHash...),
	)
	url_ := fmt.Sprintf(
		"/mms/%s/%s?direct_ip=%s&token=%s&auth=%s",
		MediaTypeStr(media.Type()),
		algo.B64UrlEnc(encFileHash),
		direct_ip,
		algo.UrlEnc(algo.B64UrlEnc(token)),
		cdn.Auth,
	)

	a.Log.Debug("host: " + `https://` + host)
	a.Log.Debug("url: " + url_)

	ua := net.UA(dev.IsBusiness, def.VERSION(dev.IsBusiness), dev.AndroidVersion, dev.Brand, dev.Model)
	_, body, e := net.HttpPost(
		host,
		url_,
		proxy, dns, dev.Ja3Config,
		cdn_http_header(host, ua),
		f_enc,
	)
	if e != nil {
		return ``, ``, e
	}
	rj, e := ajson.ParseByte(body)
	if e != nil {
		return ``, ``, errors.New(`fail parse return json: ` + string(body))
	}
	url, e := rj.Get(`url`).TryString()
	if e != nil {
		return ``, ``, errors.New(`return json missing 'url': ` + string(body))
	}
	direct_path, e := rj.Get(`direct_path`).TryString()
	if e != nil {
		return ``, ``, errors.New(`return json missing 'direct_path': ` + string(body))
	}
	return url, direct_path, e
}

func (a *Acc) set_media_conn() error {
	ch0 := &xmpp.Node{
		Tag: `media_conn`,
	}

	media_conn_id, e := a.Store.GetMediaConnId()
	if e != nil {
		return errors.Wrap(e, `get media_conn id`)
	}
	if media_conn_id != 0 {
		ch0.SetAttr(`last_id`, strconv.Itoa(int(media_conn_id)))
	}

	rn, e := a.Noise.WriteReadXmppNode(&xmpp.Node{
		Tag: `iq`,
		Attrs: []*xmpp.KeyValue{
			{Key: `id`, Value: a.Noise.NextIqId_2()},
			{Key: `to`, Value: `s.whatsapp.net`, Type: 1},
			{Key: `type`, Value: `set`},
			{Key: `xmlns`, Value: `w:m`},
		},
		Children: []*xmpp.Node{
			ch0,
		},
	})
	if e != nil {
		return e
	}

	// update
	chMc, ok := rn.FindChildByTag(`media_conn`)
	if ok { // has child `media_conn`

		mod := bson.M{}
		defer a.Store.ModifyCdn(mod)

		attrs := chMc.MapAttrs()

		// last_id
		if id_str, ok := attrs[`id`]; ok {
			if id, e := strconv.Atoi(id_str); e == nil { // `id` is number
				mod[`MediaConnId`] = uint(id)
			}
		}
		// auth
		if auth, ok := attrs[`auth`]; ok {
			mod[`Auth`] = auth
		}
		// max_buckets
		if max_buckets, ok := attrs[`max_buckets`]; ok {
			if mb, e := strconv.Atoi(max_buckets); e == nil {
				mod[`MaxBuckets`] = mb
			}
		}
		// host
		for _, chHost := range chMc.Children {
			if chHost.Tag != `host` {
				continue
			}
			attrs = chHost.MapAttrs()
			type_ := attrs[`type`]

			if type_ != `primary` {
				continue
			}
			host, ok := attrs[`hostname`]
			if !ok {
				continue
			}

			for _, h := range chHost.Children {
				if h.Tag == `download` {
					for _, t := range h.Children {
						switch t.Tag {
						case `video`:
							mod[`Video`] = host
						case `image`:
							mod[`Image`] = host
						case `ptt`:
							mod[`ppt`] = host
						case `sticker`:
							mod[`Sticker`] = host
						case `document`:
							mod[`Document`] = host
						}
					}
				}
			}
		}
	}
	return nil
}

func (c Core) DownloadMedia(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	media_t, e := j.Get(`media_type`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`invalid media_type`))
	}

	media, e := NewMedia(MediaTypeInt(media_t))
	if e != nil {
		return NewErrRet(e)
	}
	mj := j.Get(`media`)
	if media.FillFromJson(mj) != nil {
		return NewErrRet(errors.Wrap(e, `fail parse media json`))
	}

	enc, e := a.cdn_download(media)
	if e != nil {
		return NewErrRet(e)
	}

	dec, e := decryptMedia(media, enc)
	if e != nil {
		return NewErrRet(e)
	}

	ret := NewSucc()
	ret.Set(`media`, algo.B64Enc(dec))
	return ret
}

func (c Core) UploadMedia(j *ajson.Json) *ajson.Json {
	a, e := GetAccFromJson(j)
	if e != nil {
		return NewErrRet(e)
	}

	// 1. create media
	media_type, e := j.Get(`media_type`).TryString()
	if e != nil {
		return NewErrRet(errors.New(`invalid media_type`))
	}
	media, e := NewMedia(MediaTypeInt(media_type))
	if e != nil {
		return NewErrRet(e)
	}

	// 2. encrypt media binary, fill proto fields
	var f_enc []byte
	{
		// 1. get file binary
		raw, err := j.Get(media_type).TryString()
		if err != nil {
			return NewErrRet(errors.New(`missing '` + media_type + `' attribute`))
		}
		f, err := algo.B64Dec(raw)
		if err != nil {
			return NewErrRet(errors.New(`'` + media_type + `' not base64`))
		}
		// 2. fill media random mediaKey
		mediaKey := arand.Bytes(0x20)
		mediaKeyTimestamp := time.Now().Unix()

		// 3. encrypt
		f_enc, err = encryptMedia(mediaKey, media.SaltString(), f)
		if err != nil {
			return NewErrRet(errors.New(`fail encrypt content`))
		}
		j.Set(`encFileHash`, algo.B64Enc(algo.Sha256(f_enc)))
		j.Set(`fileHash`, algo.B64Enc(algo.Sha256(f)))
		j.Set(`mediaKey`, algo.B64Enc(mediaKey))
		j.Set(`mediaKeyTimestamp`, mediaKeyTimestamp)
	}

	// 3. fill media from json
	if media.FillFromJson(j) != nil {
		return NewErrRet(errors.Wrap(e, `fail parse media json`))
	}

	// 4. upload
	msg_url, dir_path, e := a.cdn_upload(media, f_enc)
	if e != nil {
		return NewErrRet(e)
	}
	// fill messageUrl/directPath
	j.Set(`messageUrl`, msg_url)
	j.Set(`directPath`, dir_path)
	if media.FillFromJson(j) != nil {
		return NewErrRet(errors.Wrap(e, `fail parse media json`))
	}

	ret := NewSucc()
	ret.Set(`media`, media.ToJson().Data())
	return ret
}
