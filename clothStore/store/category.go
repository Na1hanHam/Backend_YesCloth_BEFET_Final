package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	ImageURL     *string `json:"image_url"` // pointer = can be NULL
}

type CategoryStore struct {
	DB *pgxpool.Pool
}

func NewCategoryStore(db *pgxpool.Pool) *CategoryStore {
	return &CategoryStore{DB: db}
}

func (s *CategoryStore) List(ctx context.Context) ([]Category, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT category_id, category_name, image_url
		FROM "Categories"
		ORDER BY category_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{} // empty list -> JSON [] instead of null
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.ImageURL); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}