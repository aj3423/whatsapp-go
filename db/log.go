package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var LogLevel = ERROR

var colLog *mongo.Collection

func init() {
	colLog = client.Database(DB_NAME).Collection(`Log`)
	{
		_, e := colLog.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.M{"AccId": 1},
		})

		if e != nil {
			panic(`fail create db index`)
		}
	}
	{ // expire after 30 days
		_, e := colLog.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.M{"Time": 1},
			Options: options.Index().SetExpireAfterSeconds(3600 * 24 * 30),
		})
		if e != nil {
			panic(`fail create db index`)
		}
	}
}

const (
	DEBUG = iota
	INFO
	SUCCESS
	WARNING
	ERROR
	BUG
)

type Logger struct {
	AccId uint64
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.DoLog(DEBUG, format, args...)
}
func (l *Logger) Info(format string, args ...interface{}) {
	l.DoLog(INFO, format, args...)
}
func (l *Logger) Success(format string, args ...interface{}) {
	l.DoLog(SUCCESS, format, args...)
}
func (l *Logger) Warning(format string, args ...interface{}) {
	l.DoLog(WARNING, format, args...)
}
func (l *Logger) Error(format string, args ...interface{}) {
	l.DoLog(ERROR, format, args...)
}
func (l *Logger) Bug(format string, args ...interface{}) {
	l.DoLog(BUG, format, args...)
}
