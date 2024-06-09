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

type BookStore struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	BookName   string             `json:"name"`
	BookAuthor string             `json:"author"`
	BookISBN   string             `json:"isbn,omitempty"`
	BookPages  int                `json:"pages"`
	BookYear   int                `json:"year"`
}

func updateBookRequest(c echo.Context, coll *mongo.Collection) error {
	var updateReq BookStore
	if err := c.Bind(&updateReq); err != nil {
		log.Println("Error parsing request body:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	log.Printf("Parsed request body: %+v\n", updateReq)

	id := updateReq.ID
	if id.IsZero() {
		log.Println("Invalid book ID")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid book ID"})
	}
	log.Println("Book ID:", id)

	filter := bson.M{"_id": id}

	updateFields := bson.M{}
	if updateReq.BookName != "" {
		updateFields["BookName"] = updateReq.BookName
	}
	if updateReq.BookAuthor != "" {
		updateFields["BookAuthor"] = updateReq.BookAuthor
	}
	if updateReq.BookISBN != "" {
		updateFields["BookISBN"] = updateReq.BookISBN
	}
	if updateReq.BookPages != 0 {
		updateFields["BookPages"] = updateReq.BookPages
	}
	if updateReq.BookYear != 0 {
		updateFields["BookYear"] = updateReq.BookYear
	}

	log.Println("Update fields:", updateFields)
	update := bson.M{"$set": updateFields}

	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("Failed to update book:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update book"})
	}

	log.Println("Update result:", result)

	return c.JSON(http.StatusOK, map[string]string{"message": "Book updated successfully"})
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
	e.PUT("/api/books", func(c echo.Context) error {
		return updateBookRequest(c, coll)
	})

	e.Logger.Fatal(e.Start(":8082"))
}
