package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() (*mongo.Database, error) {
	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_DB_NAME")

	if uri == "" || dbName == "" {
		return nil, fmt.Errorf("MONGODB_URI atau MONGODB_DB_NAME kosong di .env")
	}

	clientOptions := options.Client().ApplyURI(uri)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("gagal client mongo: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("gagal ping mongo: %v", err)
	}

	log.Println("✅ Berhasil terhubung ke MongoDB")
	return client.Database(dbName), nil
}