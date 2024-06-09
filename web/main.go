package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// BookStore represents the book data model
type BookStore struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	BookName   string             `json:"name"`
	BookAuthor string             `json:"author"`
	BookISBN   string             `json:"isbn,omitempty"`
	BookPages  int                `json:"pages"`
	BookYear   int                `json:"year"`
}

// Template wraps the HTML templates
type Template struct {
	tmpl *template.Template
}

// Render renders the HTML templates
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// findAllBooks retrieves all books from the collection
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	if err != nil {
		panic(err)
	}
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"ID":         res.ID.Hex(),
			"BookName":   res.BookName,
			"BookAuthor": res.BookAuthor,
			"BookISBN":   res.BookISBN,
			"BookPages":  res.BookPages,
			"BookYear":   res.BookYear,
		})
	}
	return ret
}

// findAllAuthors retrieves all unique authors from the collection
func findAllAuthors(coll *mongo.Collection) ([]string, error) {
	var authors []string
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	for cursor.Next(context.Background()) {
		var author BookStore
		if err := cursor.Decode(&author); err != nil {
			return nil, err
		}
		authors = append(authors, author.BookAuthor)
	}
	return authors, nil
}

// findAllYears retrieves all unique publication years from the collection
func findAllYears(coll *mongo.Collection) []int {
	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	var years []int
	for _, res := range results {
		years = append(years, res.BookYear)
	}
	return years
}

// loadTemplates loads the HTML templates
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("/app/views/*.html")),
	}
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
	e.Renderer = loadTemplates()
	e.Use(middleware.Logger())

	e.Static("/css", "css")

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.Render(http.StatusOK, "index.html", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors, err := findAllAuthors(coll)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "authors.html", map[string]interface{}{
			"Authors": authors,
		})
	})

	e.GET("/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.Render(http.StatusOK, "years.html", map[string]interface{}{
			"Years": years,
		})
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(http.StatusOK, "search-bar", nil)
	})

	e.Logger.Fatal(e.Start(":8084"))
}
