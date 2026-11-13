package store

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	ImageURL     *string `json:"image_url"`
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

	categories := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.CategoryID, &c.CategoryName, &c.ImageURL); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *CategoryStore) Create(ctx context.Context, name, imageURL string) (*Category, error) {
	var c Category
	var img *string
	if strings.TrimSpace(imageURL) != "" {
		trimmed := strings.TrimSpace(imageURL)
		img = &trimmed
	}
	err := s.DB.QueryRow(ctx, `
		INSERT INTO "Categories" (category_name, image_url)
		VALUES ($1, $2)
		RETURNING category_id, category_name, image_url
	`, strings.TrimSpace(name), img).Scan(&c.CategoryID, &c.CategoryName, &c.ImageURL)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CategoryStore) Update(ctx context.Context, id int, name, imageURL string) error {
	var img *string
	if strings.TrimSpace(imageURL) != "" {
		trimmed := strings.TrimSpace(imageURL)
		img = &trimmed
	}
	_, err := s.DB.Exec(ctx, `
		UPDATE "Categories"
		SET category_name = $1, image_url = $2
		WHERE category_id = $3
	`, strings.TrimSpace(name), img, id)
	return err
}

func (s *CategoryStore) Delete(ctx context.Context, id int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM "Categories" WHERE category_id = $1`, id)
	return err
}