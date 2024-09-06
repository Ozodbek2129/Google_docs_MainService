package mongodb

import (
	"context"
	"log"
	"mainService/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDb() (*mongo.Database, error) {
	cfg := config.Load()
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Println(err)
        return nil, err
    }

	err = client.Ping(context.Background(), nil)

	if err!= nil {
        log.Println(err)
        return nil, err
    }

	return client.Database(cfg.MongoDBName), nil
} 