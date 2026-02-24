package main

import (
	"fmt"
	"log"
	"path/filepath"

	"flash-sale-be/internal/config"
	"flash-sale-be/internal/router"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	_ = godotenv.Load(".env")
	if err := godotenv.Load(filepath.Join("..", ".env")); err == nil {
	}
}

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	r := router.New(router.Deps{DB: db, Cfg: cfg})

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
