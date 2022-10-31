package handlers

import (
	"fmt"
	"goLearnGin/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RecipeHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipeHandler(ctx context.Context, collection *mongo.Collection) *RecipeHandler {
	return &RecipeHandler{
		collection: collection,
		ctx:        ctx,
	}

}
func (handler *RecipeHandler) ListRecipesHandler(c *gin.Context) {
	curr, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer curr.Close(handler.ctx)

	recipes := make([]models.Recipe, 0)
	for curr.Next(handler.ctx) {
		var recipe models.Recipe
		curr.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

func (handler *RecipeHandler) UpdateRecipeHandler(c *gin.Context) {

	id := c.Param("id")
	var recipeNew models.Recipe
	if err := c.ShouldBindJSON(&recipeNew); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, _ := primitive.ObjectIDFromHex(id)

	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{"_id": objectID}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipeNew.Name},
		{Key: "instructions", Value: recipeNew.Instructions},
		{Key: "ingredients", Value: recipeNew.Ingredients},
		{Key: "tags", Value: recipeNew.Tags},
	}}})

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "recipie updated"})
}

func (handler *RecipeHandler) DeleteRecipeHandler(c *gin.Context) {

	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)

	result, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": objectID})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Deleted": result.DeletedCount})

}

func (handler *RecipeHandler) NewRecipeHandler(c *gin.Context) {

	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting data"})
		return
	}

	c.JSON(http.StatusOK, recipe)

}
