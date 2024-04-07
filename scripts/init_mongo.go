package main

import (
	"context"
	"encoding/json"
	"goLearnGin/models"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func insertINMongo() {
	recipes := make([]models.Recipe, 0)
	file, _ := os.ReadFile("../recipes.json")

	_ = json.Unmarshal([]byte(file), &recipes)

	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongoadmin:secret@localhost:27017/?authSource=admin"))

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")

	var listOfRecipies []interface{}

	for _, recipe := range recipes {
		listOfRecipies = append(listOfRecipies, recipe)
	}

	collection := client.Database("recipiesDB").Collection("recipes")

	_, err = collection.InsertMany(ctx, listOfRecipies)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Recipes inserted into MongoDB")
}

func main() {
	insertINMongo()
}
