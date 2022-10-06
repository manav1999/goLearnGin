package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

var recipes []Recipe

func NewRecipeHandler(c *gin.Context) {
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting data"})
		return
	}

	c.JSON(http.StatusOK, recipe)

}

var ctx context.Context
var err error
var client *mongo.Client

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("conntected to db")

}

func ListRecipesHandler(c *gin.Context) {
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)
	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	id := c.Param("id")
	var recipeNew Recipe
	if err := c.ShouldBindJSON(&recipeNew); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, _ := primitive.ObjectIDFromHex(id)

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.D{{"$set", bson.D{
		{"name", recipeNew.Name},
		{"instructions", recipeNew.Instructions},
		{"ingredients", recipeNew.Ingredients},
		{"tags", recipeNew.Tags},
	}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "recipie updated"})

}

func DeleteRecipeHandler(c *gin.Context) {
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

	result, err := collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Deleted": result.DeletedCount})

}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecepies := make([]Recipe, 0)
	for i := 0; i < len(recipes); i++ {
		found := false
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecepies = append(listOfRecepies, recipes[i])
		}
	}

	c.JSON(http.StatusOK, listOfRecepies)
}

func main() {
	router := gin.Default()

	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)

	router.Run()
}
