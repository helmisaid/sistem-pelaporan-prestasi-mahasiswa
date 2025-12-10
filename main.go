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

	// Initialize Repositories
	userRepo := repository.NewUserRepository(pgDB)
	studentRepo := repository.NewStudentRepository(pgDB)
	lecturerRepo := repository.NewLecturerRepository(pgDB)
	authRepo := repository.NewAuthRepository(pgDB)
	achievementRepo := repository.NewAchievementRepository(pgDB, mongoDB)

	// Initialize Services
	studentSvc := service.NewStudentService(studentRepo)
	lecturerSvc := service.NewLecturerService(lecturerRepo)
	userSvc := service.NewUserService(userRepo, studentSvc, lecturerSvc, pgDB)
	authSvc := service.NewAuthService(authRepo)
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo, lecturerSvc)

	app := fiber.New()
	app.Use(cors.New())

	app.Static("/uploads", "./uploads")

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
	route.RegisterStudentRoutes(api, studentSvc)
	route.RegisterAchievementRoutes(api, achievementSvc)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server berjalan di port :" + port)
	log.Fatal(app.Listen(":" + port))
}