package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"stock-sim/internal/handler"
	"stock-sim/internal/repository"
	"stock-sim/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	repo := repository.NewPostgresRepository(pool)
	svc := service.NewMarketService(repo)
	router := handler.NewHandler(svc)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
