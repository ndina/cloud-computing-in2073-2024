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

type BookResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
	ISBN   string `json:"isbn,omitempty"`
}

func findAllBooksAPI(coll *mongo.Collection) ([]BookResponse, error) {
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var books []BookResponse
	for cursor.Next(context.Background()) {
		var book BookStore
		if err := cursor.Decode(&book); err != nil {
			return nil, err
		}
		bookResponse := BookResponse{
			ID:     book.ID.Hex(),
			Name:   book.BookName,
			Author: book.BookAuthor,
			Pages:  book.BookPages,
			Year:   book.BookYear,
			ISBN:   book.BookISBN,
		}
		books = append(books, bookResponse)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return books, nil
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
	e.GET("/api/books", func(c echo.Context) error {
		books, err := findAllBooksAPI(coll)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, books)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
