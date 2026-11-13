package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AdminOrder struct {
	ID       int       `json:"id"`
	Customer string    `json:"customer"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
	Date     time.Time `json:"date"`
}

type AdminReview struct {
	ID        int    `json:"id"`
	ProductID int    `json:"productId"`
	Name      string `json:"name"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
}

type AdminStore struct {
	DB *pgxpool.Pool
}

func NewAdminStore(db *pgxpool.Pool) *AdminStore {
	return &AdminStore{DB: db}
}

func (s *AdminStore) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT user_id, full_name, email, COALESCE(role, 'customer')
		FROM "User"
		ORDER BY user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []AdminUser{}
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *AdminStore) ListOrders(ctx context.Context) ([]AdminOrder, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT o.order_id, COALESCE(u.full_name, u.email), o.total_amount::float8,
		       COALESCE(o.order_status, 'pending'), o.order_date
		FROM "Orders" o
		LEFT JOIN "User" u ON u.user_id = o.user_id
		ORDER BY o.order_date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []AdminOrder{}
	for rows.Next() {
		var o AdminOrder
		if err := rows.Scan(&o.ID, &o.Customer, &o.Total, &o.Status, &o.Date); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (s *AdminStore) ListReviews(ctx context.Context) ([]AdminReview, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.review_id, r.product_id, COALESCE(u.full_name, 'Member'), r.rating, COALESCE(r.comment, '')
		FROM reviews r
		LEFT JOIN "User" u ON u.user_id = r.user_id
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []AdminReview{}
	for rows.Next() {
		var rv AdminReview
		if err := rows.Scan(&rv.ID, &rv.ProductID, &rv.Name, &rv.Rating, &rv.Comment); err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}