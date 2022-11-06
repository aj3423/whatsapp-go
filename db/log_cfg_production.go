//go:build production
// +build production

package db

import (
	"fmt"
	"time"

	"L"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	L.EnableColor(false)
}

func (l *Logger) DoLog(level int, format string, args ...interface{}) {
	if level >= LogLevel {
		bs := bson.M{
			`Time`:  time.Now(), // for TTL
			`AccId`: l.AccId,
			`Level`: level,
			`Text`:  fmt.Sprintf(format, args...),
		}
		colLog.InsertOne(ctx, bs)
	}
}
