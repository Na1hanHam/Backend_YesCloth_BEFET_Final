package store

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminUser struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	ShippingAddress string `json:"shippingAddress"`
}

type UpdateUserInput struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	ShippingAddress string `json:"shippingAddress"`
}

type AdminOrder struct {
	ID       int       `json:"id"`
	Customer string    `json:"customer"`
	Total    float64   `json:"total"`
	Status   string    `json:"status"`
	Date     time.Time `json:"date"`
}

type AdminReview struct {
	ID        int       `json:"id"`
	ProductID int       `json:"productId"`
	UserID    int       `json:"userId"`
	Name      string    `json:"name"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateUserInput struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	Role            string `json:"role"`
	ShippingAddress string `json:"shippingAddress"`
}


type AdminStore struct {
	DB *pgxpool.Pool
}

func NewAdminStore(db *pgxpool.Pool) *AdminStore {
	return &AdminStore{DB: db}
}

func (s *AdminStore) CreateUser(ctx context.Context, in CreateUserInput, passwordHash string) (*AdminUser, error) {
	role := strings.TrimSpace(in.Role)
	if role == "" {
		role = "customer"
	}
	var u AdminUser
	err := s.DB.QueryRow(ctx, `
		INSERT INTO "User" (full_name, email, password_hash, role, shipping_address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, full_name, email, COALESCE(role, 'customer'), COALESCE(shipping_address, '')
	`, strings.TrimSpace(in.Name), strings.ToLower(strings.TrimSpace(in.Email)), passwordHash, role, strings.TrimSpace(in.ShippingAddress)).Scan(
		&u.ID, &u.Name, &u.Email, &u.Role, &u.ShippingAddress,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------- Users (Members) ----------

func (s *AdminStore) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT user_id, full_name, email, COALESCE(role, 'customer'), COALESCE(shipping_address, '')
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
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.ShippingAddress); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *AdminStore) UpdateUser(ctx context.Context, userID int, in UpdateUserInput) error {
	_, err := s.DB.Exec(ctx, `
		UPDATE "User"
		SET full_name = $1,
		    email = $2,
		    role = $3,
		    shipping_address = $4,
		    updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $5
	`, strings.TrimSpace(in.Name), strings.ToLower(strings.TrimSpace(in.Email)), strings.TrimSpace(in.Role), strings.TrimSpace(in.ShippingAddress), userID)
	return err
}

func (s *AdminStore) DeleteUser(ctx context.Context, userID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM "User" WHERE user_id = $1`, userID)
	return err
}

// ---------- Orders ----------

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

// ---------- Reviews ----------

func (s *AdminStore) ListReviews(ctx context.Context) ([]AdminReview, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.review_id, r.product_id, r.user_id, COALESCE(u.full_name, 'Member'),
		       r.rating, COALESCE(r.comment, ''), r.created_at
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
		if err := rows.Scan(&rv.ID, &rv.ProductID, &rv.UserID, &rv.Name, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}

func (s *AdminStore) CreateReview(ctx context.Context, userID, productID, rating int, comment string) (*AdminReview, error) {
	var rv AdminReview
	err := s.DB.QueryRow(ctx, `
		INSERT INTO reviews (user_id, product_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING review_id, product_id, user_id, rating, COALESCE(comment, ''), created_at
	`, userID, productID, rating, strings.TrimSpace(comment)).Scan(
		&rv.ID, &rv.ProductID, &rv.UserID, &rv.Rating, &rv.Comment, &rv.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	_ = s.DB.QueryRow(ctx, `SELECT full_name FROM "User" WHERE user_id = $1`, userID).Scan(&rv.Name)
	return &rv, nil
}

func (s *AdminStore) UpdateReview(ctx context.Context, reviewID, rating int, comment string) error {
	_, err := s.DB.Exec(ctx, `
		UPDATE reviews
		SET rating = $1, comment = $2
		WHERE review_id = $3
	`, rating, strings.TrimSpace(comment), reviewID)
	return err
}

func (s *AdminStore) DeleteReview(ctx context.Context, reviewID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM reviews WHERE review_id = $1`, reviewID)
	return err
}