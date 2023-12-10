package main

import (
	"log"

	"github.com/gofiber/fiber"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author string `json:"author"`
	Title  string `json:"title"`
	Publisher string `json:"publisher"`
}
type Repository struct {
	DB *gorm.DB
}

func (r Repository)	SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database, error: %v", err)
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
