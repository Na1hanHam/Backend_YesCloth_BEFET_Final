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

	favoriteStore := store.NewFavoriteStore(pool)
	favoriteHandler := handler.NewFavoriteHandler(favoriteStore)

	adminStore := store.NewAdminStore(pool)
	adminHandler := handler.NewAdminHandler(adminStore, productStore)

		orderStore := store.NewOrderStore(pool)
	orderHandler := handler.NewOrderHandler(orderStore)


	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)

	// Categories
	mux.HandleFunc("GET /api/categories", categoryHandler.List)
	mux.HandleFunc("POST /api/admin/categories", mw.RequireAdmin(categoryHandler.Create))
	mux.HandleFunc("PUT /api/admin/categories/{id}", mw.RequireAdmin(categoryHandler.Update))
	mux.HandleFunc("DELETE /api/admin/categories/{id}", mw.RequireAdmin(categoryHandler.Delete))

	// Products
	mux.HandleFunc("GET /api/products", productHandler.List)
	mux.HandleFunc("POST /api/admin/products", mw.RequireAdmin(adminHandler.CreateProduct))
	mux.HandleFunc("PUT /api/admin/products/{id}", mw.RequireAdmin(adminHandler.UpdateProduct))
	mux.HandleFunc("DELETE /api/admin/products/{id}", mw.RequireAdmin(adminHandler.DeleteProduct))

	// Favorites (Wishlist)
	mux.HandleFunc("GET /api/favorites", mw.RequireAuth(favoriteHandler.List))
	mux.HandleFunc("POST /api/favorites", mw.RequireAuth(favoriteHandler.Add))
	mux.HandleFunc("DELETE /api/favorites/{productId}", mw.RequireAuth(favoriteHandler.Remove))

	// Reviews (Public GET, Logged-in POST, Admin PUT/DELETE)
	mux.HandleFunc("GET /api/reviews", adminHandler.ListReviews)
	mux.HandleFunc("POST /api/reviews", mw.RequireAuth(adminHandler.CreateReview))
	mux.HandleFunc("PUT /api/admin/reviews/{id}", mw.RequireAdmin(adminHandler.UpdateReview))
	mux.HandleFunc("DELETE /api/admin/reviews/{id}", mw.RequireAdmin(adminHandler.DeleteReview))

	// Admin Users & Orders
	mux.HandleFunc("GET /api/admin/users", mw.RequireAdmin(adminHandler.ListUsers))
	mux.HandleFunc("POST /api/admin/users", mw.RequireAdmin(adminHandler.CreateUser))
	mux.HandleFunc("PUT /api/admin/users/{id}", mw.RequireAdmin(adminHandler.UpdateUser))
	mux.HandleFunc("DELETE /api/admin/users/{id}", mw.RequireAdmin(adminHandler.DeleteUser))
	mux.HandleFunc("GET /api/admin/orders", mw.RequireAdmin(adminHandler.ListOrders))

		// Cart
	mux.HandleFunc("GET /api/cart", mw.RequireAuth(orderHandler.GetCart))
	mux.HandleFunc("POST /api/cart", mw.RequireAuth(orderHandler.AddToCart))
	mux.HandleFunc("PUT /api/cart", mw.RequireAuth(orderHandler.UpdateCart))

	// Orders & Checkout
	mux.HandleFunc("POST /api/checkout", mw.RequireAuth(orderHandler.Checkout))
	mux.HandleFunc("GET /api/orders/mine", mw.RequireAuth(orderHandler.ListMyOrders))

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