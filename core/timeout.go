package core

import (
	"ajson"
	"wa/def"

	"github.com/pkg/errors"
)

func (c Core) SetTimeout(j *ajson.Json) *ajson.Json {
	//a, e := GetAccFromJson(j)
	//if e != nil {
	//return NewErrRet(e)
	//}
	v, e := j.Get(`timeout`).TryInt()
	if e != nil {
		return NewErrRet(errors.Wrap(e, `wrong param "timeout"`))
	}
	def.NET_TIMEOUT = v

	return NewRet(0)
}
