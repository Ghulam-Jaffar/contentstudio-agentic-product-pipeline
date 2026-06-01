package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.Mongo.Database)

	colNames, err := db.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Collections:", colNames)

	col := db.Collection("listening_topics")

	var doc bson.M
	if err := col.FindOne(context.Background(), bson.M{"topic_id": "topic_real_brand"}).Decode(&doc); err != nil {
		log.Fatal("FindOne:", err)
	}
	fmt.Printf("Full doc: %v\n", doc)

	platforms := []string{"twitter", "instagram", "tiktok", "reddit", "threads", "facebook"}
	res, err := col.UpdateOne(
		context.Background(),
		bson.M{"_id": doc["_id"]},
		bson.M{"$set": bson.M{"enabled_platforms": platforms}},
	)
	if err != nil {
		log.Fatal("UpdateOne:", err)
	}
	fmt.Printf("Matched: %d, Modified: %d\n", res.MatchedCount, res.ModifiedCount)
}
