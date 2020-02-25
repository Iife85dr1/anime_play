package main

import (
	"context"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type AnilibriaDb struct {
	Code string
	Start int
}

type VostDB struct {
	Code int
	Start int
}

func mongoConnect() *mongo.Client {
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil{
		log.Fatalln(err)
	}
	// Create connect
	err = mongoClient.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return mongoClient
}
