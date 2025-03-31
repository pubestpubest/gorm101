package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db := gormSetup()
	db.AutoMigrate(&Book{}, &User{})
	fmt.Println("Migration completed")

	app := fiber.New()
	app.Post("/register", func(c *fiber.Ctx) error {
		return createUser(db, c)
	})
	app.Post("/login", func(c *fiber.Ctx) error {
		return loginUser(db, c)
	})
	app.Use("/book", authRequired)
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

func createUser(db *gorm.DB, c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	db.Create(&user)
	return c.JSON(user)
}

func loginUser(db *gorm.DB, c *fiber.Ctx) error {
	var input User
	var user User
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	//find user
	db.Where("email = ?", input.Email).First(&user)
	//Check hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	//create jwt
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    t,
		Expires:  time.Now().Add(time.Hour * 72),
		HTTPOnly: true,
	})
	return c.JSON(fiber.Map{"message": "success"})
}

func authRequired(c *fiber.Ctx) error {
	cookie := c.Cookies("jwt")
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.SendStatus(fiber.StatusUnauthorized)
	}
	return c.Next()
}
