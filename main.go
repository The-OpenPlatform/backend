package main

import (
	"database/sql"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	var err error

	connStr := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	app := fiber.New()
	app.Use(cors.New())

	const apiPrefix = "/api" // Kept here for testing in case the stripping out of /api/v1 does not work or similar

	app.Get(apiPrefix+"/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the root endpoint!")
	})

	app.Get(apiPrefix+"/hello", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get(apiPrefix+"/message", getRandomMessage)

	app.Get(apiPrefix+"/status", getStatus)

	log.Println("Server is running on port 3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

func getStatus(c *fiber.Ctx) error {
	podName, err := os.Hostname()
	if err != nil {
		podName = "unknown"
	}

	return c.SendString(podName)
}

func getRandomMessage(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT message FROM messages")
	if err != nil {
		return c.Status(500).SendString("Failed to query messages")
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return c.Status(500).SendString("Error reading message")
		}
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		return c.Status(404).SendString("No messages found")
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := rand.Intn(len(messages))
	return c.SendString(messages[randomIndex])
}
