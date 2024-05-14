package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
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

type BookRequest struct {
	Name   string `json:"name"`
	Author string `json:"author"`
	Pages  int    `json:"pages"`
	Year   int    `json:"year"`
	ISBN   string `json:"isbn,omitempty"`
}

// Wraps the "Template" struct to associate a necessary method
// to determine the rendering procedure
type Template struct {
	tmpl *template.Template
}

// Preload the available templates for the view folder.
// This builds a local "database" of all available "blocks"
// to render upon request, i.e., replace the respective
// variable or expression.
// For more on templating, visit https://jinja.palletsprojects.com/en/3.0.x/templates/
// to get to know more about templating
// You can also read Golang's documentation on their templating
// https://pkg.go.dev/text/template
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("/app/views/*.html")),
	}
}

// Method definition of the required "Render" to be passed for the Rendering
// engine.
// Contraire to method declaration, such syntax defines methods for a given
// struct. "Interfaces" and "structs" can have methods associated with it.
// The difference lies that interfaces declare methods whether struct only
// implement them, i.e., only define them. Such differentiation is important
// for a compiler to ensure types provide implementations of such methods.
func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// Here we make sure the connection to the database is correct and initial
// configurations exists. Otherwise, we create the proper database and collection
// we will store the data.
// To ensure correct management of the collection, we create a return a
// reference to the collection to always be used. Make sure if you create other
// files, that you pass the proper value to ensure communication with the
// database
// More on what bson means: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
func prepareDatabase(client *mongo.Client, dbName string, collecName string) (*mongo.Collection, error) {
	db := client.Database(dbName)

	names, err := db.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, collecName) {
		cmd := bson.D{{"create", collecName}}
		var result bson.M
		if err = db.RunCommand(context.TODO(), cmd).Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	coll := db.Collection(collecName)
	return coll, nil
}

// Here we prepare some fictional data and we insert it into the database
// the first time we connect to it. Otherwise, we check if it already exists.
func prepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{
			BookName:   "The Vortex",
			BookAuthor: "José Eustasio Rivera",
			BookISBN:   "958-30-0804-4",
			BookPages:  292,
			BookYear:   1924,
		},
		{
			BookName:   "Frankenstein",
			BookAuthor: "Mary Shelley",
			BookISBN:   "978-3-649-64609-9",
			BookPages:  280,
			BookYear:   1818,
		},
		{
			BookName:   "The Black Cat",
			BookAuthor: "Edgar Allan Poe",
			BookISBN:   "978-3-99168-238-7",
			BookPages:  280,
			BookYear:   1843,
		},
	}

	// This syntax helps us iterate over arrays. It behaves similar to Python
	// However, range always returns a tuple: (idx, elem). You can ignore the idx
	// by using _.
	// In the topic of function returns: sadly, there is no standard on return types from function. Most functions
	// return a tuple with (res, err), but this is not granted. Some functions
	// might return a ret value that includes res and the err, others might have
	// an out parameter.
	for _, book := range startData {
		cursor, err := coll.Find(context.TODO(), book)
		var results []BookStore
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		if len(results) > 1 {
			log.Fatal("more records were found")
		} else if len(results) == 0 {
			result, err := coll.InsertOne(context.TODO(), book)
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", result)
			}

		} else {
			for _, res := range results {
				cursor.Decode(&res)
				fmt.Printf("%+v\n", res)
			}
		}
	}
}

// Generic method to perform "SELECT * FROM BOOKS" (if this was SQL, which
// it is not :D ), and then we convert it into an array of map. In Golang, you
// define a map by writing map[<key type>]<value type>{<key>:<value>}.
// interface{} is a special type in Golang, basically a wildcard...
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
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
		})
	}

	return ret
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

func updateBookRequest(c echo.Context, coll *mongo.Collection) error {
    var updateReq BookStore
    if err := c.Bind(&updateReq); err != nil {
        log.Println("Error parsing request body:", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
    }

    log.Printf("Parsed request body: %+v\n", updateReq)
	log.Printf("Parsed request body 2: %+v\n", updateReq.ID, updateReq.BookName, updateReq.BookPages)

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

	uri := os.Getenv("DATABASE_URI")
	if len(uri) == 0 {
		fmt.Printf("failure to load env variable\n")
		os.Exit(1)
	}

	// TODO: make sure to pass the proper username, password, and port
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

	// This is another way to specify the call of a function. You can define inline
	// functions (or anonymous functions, similar to the behavior in Python)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// You can use such name for the database and collection, or come up with
	// one by yourself!
	coll, err := prepareDatabase(client, "exercise-1", "information")

	prepareData(client, coll)

	// Here we prepare the server
	e := echo.New()

	// Define our custom renderer
	e.Renderer = loadTemplates()

	// Log the requests. Please have a look at echo's documentation on more
	// middleware
	e.Use(middleware.Logger())

	e.Static("/css", "css")

	// Endpoint definition. Here, we divided into two groups: top-level routes
	// starting with /, which usually serve webpages. For our RESTful endpoints,
	// we prefix the route with /api to indicate more information or resources
	// are available under such route.
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.Render(200, "book-table", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors, err := findAllAuthors(coll)
		if err != nil {
			return err
		}
		// Render the authors block in index.html template and pass authors data
		return c.Render(http.StatusOK, "authors", map[string]interface{}{
			"Authors": authors,
		})
	})
	
	e.GET("/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.Render(http.StatusOK, "years", map[string]interface{}{
			"Years": years,
		})
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(200, "search-bar", nil)
	})

	e.GET("/create", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.GET("/api/books", func(c echo.Context) error {
		books, err := findAllBooksAPI(coll)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, books)
	})

	e.POST("/api/books", func(c echo.Context) error {
        return createBookRequest(c, coll)
    })

	e.PUT("/api/books", func(c echo.Context) error {
		return updateBookRequest(c, coll)
	})

	e.DELETE("/api/books/:id", func(c echo.Context) error {
		return deleteBookRequest(c, coll)
	})


	e.Logger.Fatal(e.Start(":3030"))
}
