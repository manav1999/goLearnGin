package handlers

import (
	"context"
	"crypto/sha256"
	"goLearnGin/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JoinVerse/xid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

type Claims struct {
	Username string `json:"username:"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
		c.AbortWithStatus(401)
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h := sha256.New()

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"username": user.UserName,
		"password": string(h.Sum([]byte(user.Password))),
	})

	if cur.Err() != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or Password"})
		return
	}

	expirationTime := time.Now().Add(10 * time.Minute)

	claims := &Claims{
		Username:       user.UserName,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime),},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		c.JSON(
			http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	JWTOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}


	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username",user.UserName)
	session.Set("token",sessionToken)
	session.Save()


	c.JSON(http.StatusOK, JWTOutput)

}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
			c.AbortWithStatus(401)
		}
		tokenValue := c.GetHeader("Authorization")
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			log.Print(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		if !tkn.Valid {
			log.Print(tkn.Claims.Valid().Error())
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}

func (handler *AuthHandler) SignUpHandler(c *gin.Context) {
	if c.GetHeader("X-API-KEY") != os.Getenv("X_API_KEY") {
		c.AbortWithStatus(401)
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h := sha256.New()

	_, err := handler.collection.InsertOne(handler.ctx, bson.M{"username": user.UserName, "password": string(h.Sum([]byte(user.Password)))})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user.UserName)
}

func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}

	if !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
		return
	}

	if time.Until(claims.ExpiresAt.Time) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token Not Expired Yet"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(os.Getenv("JWT_SECRET"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erro": err.Error()})
		return
	}

	JWTOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}

	c.JSON(http.StatusOK, JWTOutput)

}
