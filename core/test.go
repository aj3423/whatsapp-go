package core

import (
	"ajson"
)

func (c Core) Test(j *ajson.Json) *ajson.Json {
	a, _ := GetAccFromJson(j)
	_ = a

	return NewRet(0)
}
