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

	_ "sistem-pelaporan-prestasi-mahasiswa/docs"

	"github.com/gofiber/fiber/v2/middleware/logger"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title Sistem Pelaporan Prestasi Mahasiswa API
// @version 1.0
// @description API untuk mengelola data prestasi mahasiswa, dosen wali, dan laporan.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: File .env tidak ditemukan, menggunakan system environment")
	}

	pgDB, err := database.ConnectPostgres()
	if err != nil {
		log.Fatal("❌ Gagal konek PostgreSQL: ", err)
	}
	defer pgDB.Close()

	mongoDB, err := database.ConnectMongo()
	if err != nil {
		log.Fatal("❌ Gagal konek MongoDB: ", err)
	}

	userRepo := repository.NewUserRepository(pgDB)
	studentRepo := repository.NewStudentRepository(pgDB)
	lecturerRepo := repository.NewLecturerRepository(pgDB)
	authRepo := repository.NewAuthRepository(pgDB)
	achievementRepo := repository.NewAchievementRepository(pgDB, mongoDB)
	reportRepo := repository.NewReportRepository(pgDB, mongoDB)


	lecturerSvc := service.NewLecturerService(lecturerRepo)
	studentSvc := service.NewStudentService(studentRepo, lecturerSvc)
	userSvc := service.NewUserService(userRepo, studentSvc, lecturerSvc, pgDB)
	authSvc := service.NewAuthService(authRepo)
	achievementSvc := service.NewAchievementService(achievementRepo, studentRepo, lecturerSvc)
	reportSvc := service.NewReportService(reportRepo, studentRepo, lecturerSvc)

	app := fiber.New()
	app.Use(cors.New())
	app.Use(logger.New())

	app.Static("/uploads", "./uploads")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Server berjalan dengan koneksi Hybrid (Postgres + Mongo)",
			"status":  "OK",
		})
	})

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	api := app.Group("/api/v1")

	route.RegisterAuthRoutes(api, authSvc)
	route.RegisterUserRoutes(api, userSvc)
	route.RegisterStudentRoutes(api, studentSvc, achievementSvc)
	route.RegisterLecturerRoutes(api, lecturerSvc)
	route.RegisterAchievementRoutes(api, achievementSvc)
	route.RegisterReportRoutes(api, reportSvc)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server berjalan di port :" + port)
	log.Println("📄 Swagger UI tersedia di http://localhost:" + port + "/swagger/index.html")
	log.Fatal(app.Listen(":" + port))
}