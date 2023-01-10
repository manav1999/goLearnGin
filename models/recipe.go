package models

import (
	
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//recipe model 
type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

//user model 
type User struct {
	Password string `json:"password"`
	UserName string `json:"username"`
}