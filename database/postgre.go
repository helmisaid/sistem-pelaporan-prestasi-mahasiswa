package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" 
)

func ConnectPostgres() (*sql.DB, error) {
	dsn := os.Getenv("POSTGRES_DSN") 
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_DSN tidak ditemukan di .env")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka driver postgres: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping database: %v", err)
	}

	log.Println("✅ Berhasil terhubung ke database PostgreSQL")
	return db, nil
}