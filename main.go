package main

import (
	"log"
	"os"

	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/app/service"
	"sistem-pelaporan-prestasi-mahasiswa/database"
	"sistem-pelaporan-prestasi-mahasiswa/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Environment
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: File .env tidak ditemukan, menggunakan system environment")
	}

	// Koneksi PostgreSQL
	pgDB, err := database.ConnectPostgres()
	if err != nil {
		log.Fatal("❌ Gagal konek PostgreSQL: ", err)
	}
	defer pgDB.Close()

	// Koneksi MongoDB 
	mongoDB, err := database.ConnectMongo()
	if err != nil {
		log.Fatal("❌ Gagal konek MongoDB: ", err)
	}
	// Placeholder
	_ = mongoDB

	// Repository 
	authRepo := repository.NewAuthRepository(pgDB)
	userRepo := repository.NewUserRepository(pgDB)

	// Service 
	authSvc := service.NewAuthService(authRepo)
	userSvc := service.NewUserService(userRepo, pgDB)

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Server berjalan dengan koneksi Hybrid (Postgres + Mongo)",
			"status":  "OK",
		})
	})

	// Grouping Route API 
	api := app.Group("/api/v1")

	route.RegisterAuthRoutes(api, authSvc)
	route.RegisterUserRoutes(api, userSvc)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server berjalan di port :" + port)
	log.Fatal(app.Listen(":" + port))
}