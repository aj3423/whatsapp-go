package core

import (
	"ajson"
	"wa/def"
)

func (c Core) Version(j *ajson.Json) *ajson.Json {
	x := ajson.New()
	x.Set("biz", def.VERSION_biz)
	x.Set("personal", def.VERSION_psnl)

	return NewJsonRet(x)
}
