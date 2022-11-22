package main

import (
	"context"
	"fmt"
	"goLearnGin/handlers"
	"log"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipeHandler *handlers.RecipeHandler
var authHandler *handlers.AuthHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MONGODB")
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := redisClient.Ping(ctx)
	fmt.Println(status)

	recipeHandler = handlers.NewRecipeHandler(ctx, collection, redisClient)
	authHandler = &handlers.AuthHandler{}
}


func main() {
	router := gin.Default()
	authorised := router.Group("/")
	authorised.Use(authHandler.AuthMiddleware())
	router.GET("/recipes", recipeHandler.ListRecipesHandler)
	router.POST("/signin",authHandler.SignInHandler)
	authorised.POST("/recipes", recipeHandler.NewRecipeHandler)
	authorised.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
	authorised.DELETE("/recipes/:id", recipeHandler.DeleteRecipeHandler)
	router.Run()
}
