package core

import (
	"ajson"

	"github.com/pkg/errors"
)

func NewRet(ec int) *ajson.Json {
	j := ajson.New()
	j.Set("ErrCode", ec)
	return j
}
func NewSucc() *ajson.Json {
	return NewRet(0)
}
func NewErrRet(e error) *ajson.Json {
	j := ajson.New()
	j.Set("ErrCode", 3423)
	j.Set("ErrMsg", e.Error())
	return j
}
func NewCrashRet() *ajson.Json {
	return NewErrRet(errors.New("Crashed..."))
}

func NewJsonRet(result *ajson.Json) *ajson.Json {
	j := NewSucc()
	j.Set(`Result`, result)
	return j
}
