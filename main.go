package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db := gormSetup()
	db.AutoMigrate(&Book{})
	fmt.Println("Migration completed")

	app := fiber.New()

	app.Get("/book", func(c *fiber.Ctx) error {
		return c.JSON(GetBooks(db))
	})
	app.Get("/book/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.JSON(GetBook(db, uint(id)))
	})
	app.Post("/book", func(c *fiber.Ctx) error {
		var book Book
		err := c.BodyParser(&book)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		CreateBook(db, &book)
		return c.SendString("Book created successfully")
	})
	app.Put("/book/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		var book Book
		err = c.BodyParser(&book)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		book.ID = uint(id)
		UpdateBook(db, &book)
		return c.SendString("Book updated successfully")
	})
	app.Delete("book/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		DeleteBook(db, uint(id))
		return c.SendString("Book deleted successfully")
	})

	app.Listen(":4040")
}

func gormSetup() *gorm.DB {
	//Load environment variable from .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	host := "localhost"
	port := 5432
	dbuser := os.Getenv("POSTGRES_USER")
	dbpass := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

	//Make connection string
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, dbuser, dbpass, dbname)

	//Init GORM database instance
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get the underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database!")
	return db
}
