package main

import (
	"log"
	"os"
	"sistem-pelaporan-prestasi-mahasiswa/database" 

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load Environment
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: File .env tidak ditemukan, menggunakan system environment")
	}

	// Koneksi PostgreSQL 
	pgDB, err := database.ConnectPostgres()
	if err != nil {
		log.Fatal("❌ Gagal konek PostgreSQL: ", err)
	}
	defer pgDB.Close() 

	// Setup Koneksi MongoDB
	mongoDB, err := database.ConnectMongo()
	if err != nil {
		log.Fatal("❌ Gagal konek MongoDB: ", err)
	}

	_ = mongoDB 

	// Inisialisasi Fiber
	app := fiber.New()
	app.Use(cors.New())

	// Route Test
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Server berjalan dengan koneksi Hybrid (Postgres + Mongo)",
			"status":  "OK",
		})
	})

	// run server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	
	log.Println("🚀 Server berjalan di port :" + port)
	log.Fatal(app.Listen(":" + port))
}