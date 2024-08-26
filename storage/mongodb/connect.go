package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDb() (*mongo.Database, error) {
	client, err := mongo.Connect(context.Background(), options.Client().
        ApplyURI("mongodb://localhost:27017").SetAuth(options.Credential{Username: "root", Password: "example"}))
	if err != nil {
		log.Println(err)
        return nil, err
    }

	err = client.Ping(context.Background(), nil)

	if err!= nil {
        log.Println(err)
        return nil, err
    }

	return client.Database("google_docs"), nil
} 