package main

import (

	"context"
	"time"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client		*mongo.Client
	collection	*mongo.Collection
)

func search(filter bson.M) {

	var result struct {
		Value float64
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func query(target bson.D) {

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cur, err := collection.Find(ctx, target)
	if err != nil { log.Fatal(err) }
	defer cur.Close(ctx)
	for cur.Next(ctx) {
   		var result bson.M
   		err := cur.Decode(&result)
   		if err != nil { log.Fatal(err) }
   		fmt.Println(result)
	}
	if err := cur.Err(); err != nil {
  		log.Fatal(err)
	}
}

func insert(col string) {
	
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	collection = client.Database("testing").Collection(col)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	id := res.InsertedID
	fmt.Println(id)
}

func main() {

	insert("numbers")
	query(bson.D{})
	search(bson.M{"name": "pi"})
}