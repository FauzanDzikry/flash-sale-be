package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"flash-sale-be/internal/config"
	"flash-sale-be/internal/queue"
	"flash-sale-be/internal/repository"
	"flash-sale-be/internal/router"
	"flash-sale-be/internal/service"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis: %v", err)
	}

	q := queue.NewRedisQueue(rdb)
	checkoutRepo := repository.NewCheckoutRepository(db)
	productsRepo := repository.NewProductsRepository(db)
	checkoutSvc := service.NewCheckoutService(checkoutRepo, productsRepo, q, db)

	r := router.New(router.Deps{
		DB:              db,
		Cfg:             cfg,
		CheckoutService: checkoutSvc,
		Redis:           rdb,
	})

	addr := ":8080"
	log.Printf("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
