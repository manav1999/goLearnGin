package main

import (
	"context"
	"fmt"
	"goLearnGin/handlers"
	"log"
	"os"

	redisStore "github.com/gin-contrib/sessions/redis"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Contains functions to handle GET,POST,PUT,DELETE request
var recipeHandler *handlers.RecipeHandler

// Contains functions to handle Authentication
var authHandler *handlers.AuthHandler

// Initalises mongodb, redis and then creates recipie handler and authhandler
func init() {

	ctx := context.Background()
	//Creates a mongodb client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MONGODB")
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	//Creates a redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := redisClient.Ping(ctx)
	fmt.Println(status)

	recipeHandler = handlers.NewRecipeHandler(ctx, collection, redisClient)
	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	//Main router
	router := gin.Default()
	//Authentication Middleware, allows request to execute further only if user is authenticated
	authorised := router.Group("/")
	authorised.Use(authHandler.AuthMiddleware())
	//GET reciepes
	router.GET("/recipes", recipeHandler.ListRecipesHandler)
	//Used to get token for an already created user
	router.POST("/signin", authHandler.SignInHandler)
	//Used to create user
	router.POST("/signup", authHandler.SignUpHandler)
	//used to generate refesh token
	authorised.POST("/refresh", authHandler.RefreshHandler)

	//Creates a new recipe
	authorised.POST("/recipes", recipeHandler.NewRecipeHandler)
	//Updates the existing recipe
	authorised.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
	//Used to delete recipe
	authorised.DELETE("/recipes/:id", recipeHandler.DeleteRecipeHandler)

	//Session Manager

	store, _ := redisStore.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))

	router.Use(sessions.Sessions("recipes_api", store))

	router.Run()
}
