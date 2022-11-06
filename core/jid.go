package core

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// 8613011112222.0:1@s.whatsapp.net -> 9613011112222@s.whatsapp.net
func clear_jid_device(jid string) string {
	p := strings.Index(jid, ":")
	if p > 0 {
		return jid[0:p-2] + jid[p+2:]
	}
	return jid
}

// 8613322222222.0:1@s.whatsapp.net -> 8613322222222  1
func split_jid(jid string) (recid uint64, devid uint32, e error) {
	jid = strings.ReplaceAll(jid, "@s.whatsapp.net", "")

	{ // recid
		v := strings.Split(jid, ".")
		if len(v) == 0 {
			e = errors.New("invalid jid: " + jid)
			return
		}

		var x int
		x, e = strconv.Atoi(v[0])
		if e != nil {
			return
		}
		recid = uint64(x)
	}
	{ // devid
		v := strings.Split(jid, ":")
		if len(v) == 2 {
			var x int
			x, e = strconv.Atoi(v[1])
			if e != nil {
				return
			}
			devid = uint32(x)
		}
	}
	return
}
