package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type HistoricalFigure struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Birthdate string             `bson:"birthdate" json:"birthdate"`
	Fields    []string           `bson:"fields" json:"fields"`
	Bio       string             `bson:"bio" json:"bio"`
	ImageURL  string             `bson:"image_url" json:"image_url"`
	SourceURL string             `bson:"source_url" json:"source_url"`
}

var collection *mongo.Collection

// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatal("❌ Error loading .env file:", err)
// 	}

// 	uri := os.Getenv("MONGO_DB_ATLAS_URI")
// 	dbName := os.Getenv("MONGO_DB_NAME")
// 	collName := os.Getenv("MONGODB_COLLECTION")

// 	if uri == "" || dbName == "" || collName == "" {
// 		log.Fatal("❌ Environment variables are missing")
// 	}

// 	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
// 	if err != nil {
// 		log.Fatal("❌ Mongo client creation failed:", err)
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	if err := client.Connect(ctx); err != nil {
// 		log.Fatal("❌ Mongo client connect failed:", err)
// 	}

// 	collection = client.Database(dbName).Collection(collName)
// 	log.Println("✅ Connected to MongoDB Atlas")
// }

func init() {
	// Only load .env file locally
	if os.Getenv("GIN_MODE") != "release" {
		if err := godotenv.Load(); err != nil {
			log.Println("⚠️ Could not load .env file (this is fine in production)")
		}
	}

	uri := os.Getenv("MONGO_DB_ATLAS_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	collName := os.Getenv("MONGODB_COLLECTION")

	if uri == "" || dbName == "" || collName == "" {
		log.Fatal("❌ Environment variables are missing")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("❌ Mongo client creation failed:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatal("❌ Mongo client connect failed:", err)
	}

	collection = client.Database(dbName).Collection(collName)
	log.Println("✅ Connected to MongoDB Atlas")
}

func main() {

	r := gin.Default()

	// Manual CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/figures", getAllFigures)
	r.GET("/figures/:id", getFigureByID)
	r.POST("/figures", createFigure)
	r.PUT("/figures/:id", updateFigure)
	r.DELETE("/figures/:id", deleteFigure)

	// Handle 404s with CORS
	r.NoRoute(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		c.JSON(404, gin.H{"error": "Not found"})
	})

	log.Println("🚀 Server running on http://0.0.0.0:8080")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local
	}
	r.Run(":" + port)
	// r.Run("0.0.0.0:8080")

}

func getAllFigures(c *gin.Context) {
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data"})
		return
	}
	defer cursor.Close(context.TODO())

	var figures []HistoricalFigure
	if err := cursor.All(context.TODO(), &figures); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding data"})
		return
	}

	c.JSON(http.StatusOK, figures)
}

func getFigureByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var figure HistoricalFigure
	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&figure)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Figure not found"})
		return
	}

	c.JSON(http.StatusOK, figure)
}

func createFigure(c *gin.Context) {
	var figure HistoricalFigure
	if err := c.ShouldBindJSON(&figure); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	result, err := collection.InsertOne(context.TODO(), figure)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}

	figure.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, figure)
}

func updateFigure(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var figure HistoricalFigure
	if err := c.ShouldBindJSON(&figure); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	update := bson.M{
		"$set": figure,
	}

	_, err = collection.UpdateByID(context.TODO(), objID, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	figure.ID = objID
	c.JSON(http.StatusOK, figure)
}

func deleteFigure(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Figure deleted"})
}
