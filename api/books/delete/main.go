package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func deleteBookRequest(c echo.Context, coll *mongo.Collection) error {
	id := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}

	filter := bson.M{"_id": objID}

	result, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete book"})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Book deleted successfully"})
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_URI")
	if len(uri) == 0 {
		fmt.Printf("failure to load env variable\n")
		os.Exit(1)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Printf("failed to create client for MongoDB\n")
		os.Exit(1)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Printf("failed to connect to MongoDB, please make sure the database is running\n")
		os.Exit(1)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("exercise-1").Collection("information")

	e := echo.New()
	e.DELETE("/api/books/:id", func(c echo.Context) error {
		return deleteBookRequest(c, coll)
	})

	e.Logger.Fatal(e.Start(":8083"))
}
