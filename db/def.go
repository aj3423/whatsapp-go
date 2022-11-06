package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

var ctx = context.Background()

func init() {
	var e error
	client, e = mongo.Connect(
		ctx, options.Client().ApplyURI("mongodb://127.0.0.1"))
	if e != nil {
		panic(e)
	}
}
