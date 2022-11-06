//go:build !production
// +build !production

package db

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"go.mongodb.org/mongo-driver/bson"
)

var writer io.Writer

func init() {
	writer = io.MultiWriter(colorable.NewColorableStdout())
	log.SetOutput(writer)
}

// run on my self kali
func (l *Logger) DoLog(level int, format string, args ...interface{}) {
	if level >= LogLevel {
		switch level {
		case DEBUG:
			log.Printf(color.HiWhiteString(format), args...)
		case INFO:
			log.Printf(color.HiBlueString(format), args...)
		case SUCCESS:
			log.Printf(color.HiGreenString(format), args...)
		case WARNING:
			log.Printf(color.HiYellowString(format), args...)
		case ERROR:
			log.Printf(color.HiRedString(format), args...)
		case BUG:
			log.Printf(color.HiRedString(format), args...)
		}
	}
	// log to db
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
