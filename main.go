package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db := gormSetup()
	db.AutoMigrate(&Book{})
	fmt.Println("Migration completed")
	//CreateBook(db, &Book{Name: "Go102", Author: "Pubest", Page: 101})
	// cbook := GetBook(db, 1)
	// cbook.Page = 200
	// UpdateBook(db, cbook)
	// fmt.Println(GetBooks(db))
	DeleteBook(db, 2)
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

func CreateBook(db *gorm.DB, book *Book) {
	result := db.Create(book)
	if result.Error != nil {
		log.Fatalf("Error creating book: %v", result.Error)
	}
	fmt.Println("Book created successfully")
}

func UpdateBook(db *gorm.DB, book *Book) {
	result := db.Save(&book)
	if result.Error != nil {
		log.Fatalf("Error update book: %v", result.Error)
	}
	fmt.Println("Book updated successfully")
}

func GetBook(db *gorm.DB, id uint) *Book {
	var book Book
	result := db.First(&book, id)
	if result.Error != nil {
		log.Fatalf("Error finding book: %v", result.Error)
	}
	return &book
}

func GetBooks(db *gorm.DB) *[]Book {
	var books []Book
	result := db.Find(&books)
	if result.Error != nil {
		log.Fatalf("Error finding books: %v", result.Error)
	}
	return &books
}

func DeleteBook(db *gorm.DB, id uint) {
	var book Book
	result := db.Unscoped().Delete(&book, id)
	if result.Error != nil {
		log.Fatalf("Error deleting books: %v", result.Error)
	}
}
