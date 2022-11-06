package net

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	fhttp "github.com/useflyent/fhttp"

	utls "github.com/refraction-networking/utls"

	"wa/def"
)

func UA(is_biz bool, ver, android_ver, brand, model string) string {
	smba := "Android"
	if is_biz {
		smba = "SMBA"
	}
	ua := fmt.Sprintf(
		"WhatsApp/%s %s/%s Device/%s-%s",
		ver,
		smba,
		android_ver,
		brand,
		strings.ReplaceAll(model, " ", "_"),
	)

	return ua
}

func HttpReq(
	method, host, url_,
	proxyAddr string, dns map[string]string, ja3_str string,
	hdr fhttp.Header,
	post_body []byte,
) (int, []byte, error) {

	if ja3_str == `` {
		ja3_str = def.Ja3_WhiteMi6x
	}
	var ja3_transport *fhttp.Transport
	var e error

	if len(proxyAddr) > 0 {
		ja3_transport, e = NewTransportWithConfigAndSocks5(
			ja3_str, proxyAddr, &utls.Config{
				InsecureSkipVerify: true,
				MaxVersion:         tls.VersionTLS13,
			})
	} else {
		ja3_transport, e = NewTransportWithConfig(
			ja3_str, &utls.Config{
				InsecureSkipVerify: true,
				MaxVersion:         tls.VersionTLS13,
			})
	}
	if e != nil {
		return 0, nil, e
	}

	client := &fhttp.Client{
		Transport: ja3_transport,
	}

	// timeout
	ctx, cancel := context.WithCancel(context.TODO())
	timer := time.AfterFunc(time.Duration(def.NET_TIMEOUT)*time.Second, cancel)
	defer timer.Stop()

	host_modified := host

	if dns != nil {
		ip, ok := dns[host]
		if ok && len(ip) > 0 {
			host_modified = ip
		}
	}

	// for debugging
	//req, e := fhttp.NewRequest(method, "http://"+host_modified+url_, bytes.NewReader(post_body))
	req, e := fhttp.NewRequest(method, "https://"+host_modified+url_, bytes.NewReader(post_body))
	if e != nil {
		return 0, nil, e
	}
	req = req.WithContext(ctx)

	// force http header.Host, if changed by dns
	//if host != host_modified {
	//req.Host = host
	//}

	req.Header = hdr

	resp, e := client.Do(req)
	if e != nil {
		return 0, nil, e
	}

	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return 0, nil, e
	}

	return resp.StatusCode, body, nil
}

func HttpGet(host, url_, proxyAddr string, dns map[string]string, ja3_str string, hdr fhttp.Header) (int, []byte, error) {
	return HttpReq(`GET`, host, url_, proxyAddr, dns, ja3_str, hdr, nil)
}
func HttpPost(host, url_, proxyAddr string, dns map[string]string, ja3_str string, hdr fhttp.Header, body []byte) (int, []byte, error) {
	return HttpReq(`POST`, host, url_, proxyAddr, dns, ja3_str, hdr, body)
}
