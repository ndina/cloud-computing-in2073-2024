package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
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

type BookRequest struct {
	Name   string `json:"name"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
	ISBN   string `json:"isbn,omitempty"`
}

func createBookRequest(c echo.Context, coll *mongo.Collection) error {
	fmt.Println("Received POST request to create a new book")
	var bookReq BookRequest
	if err := c.Bind(&bookReq); err != nil {
		fmt.Println("Error parsing request body:", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}
	fmt.Println("Parsed request body:", bookReq)

	book := BookStore{
		BookName:   bookReq.Name,
		BookAuthor: bookReq.Author,
		BookPages:  bookReq.Pages,
		BookYear:   bookReq.Year,
		BookISBN:   bookReq.ISBN,
	}
	_, err := coll.InsertOne(context.Background(), book)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add book"})
	}

	return c.NoContent(http.StatusOK)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("DATABASE_URI")
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
	e.POST("/api/books", func(c echo.Context) error {
		return createBookRequest(c, coll)
	})

	e.Logger.Fatal(e.Start(":8081"))
}
