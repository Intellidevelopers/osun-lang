package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

// InitMongoFromEnv connects to MongoDB if MONGO_URI is set.
// Example: mongodb://localhost:27017/osun
func InitMongoFromEnv() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Mongo connect error:", err)
		return
	}
	if err := client.Ping(ctx, nil); err != nil {
		fmt.Println("Mongo ping error:", err)
		_ = client.Disconnect(ctx)
		return
	}
	mongoClient = client
	fmt.Println("Mongo connected")
}

// DBInsertMongo inserts a JSON object into the named collection in database "osun".
func DBInsertMongo(collection string, jsonStr string) error {
	if mongoClient == nil {
		return fmt.Errorf("mongo not configured")
	}
	var doc map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := mongoClient.Database("osun").Collection(collection).InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("mongo insert error: %w", err)
	}
	return nil
}
