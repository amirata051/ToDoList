package db

import (
	"context"
	"log"

	"github.com/akhilsharma/todo/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Init initializes and returns a new MongoDB database connection.
func Init() *mongo.Database {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.HostName))
	if err != nil {
		log.Fatal(err)
	}

	// Ping the primary
	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return client.Database(config.DbName)
}