package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"clothStore/handler"
	"clothStore/mw"
	"clothStore/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	// 1. Connect to Database
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Connected to PostgreSQL")

	// 2. Wire Store & Handlers
	userStore := store.NewUserStore(pool)
	authHandler := handler.NewAuthHandler(userStore)

	categoryStore := store.NewCategoryStore(pool)
	categoryHandler := handler.NewCategoryHandler(categoryStore)

	productStore := store.NewProductStore(pool)
	productHandler := handler.NewProductHandler(productStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/categories", categoryHandler.List)
	mux.HandleFunc("GET /api/products", productHandler.List)
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// 3. Accept Frontend Requests
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mw.CORS(mux)); err != nil {
		log.Fatal(err)
	}
}