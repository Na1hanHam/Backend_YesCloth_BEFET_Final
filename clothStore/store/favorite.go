package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FavoriteStore struct {
	DB *pgxpool.Pool
}

func NewFavoriteStore(db *pgxpool.Pool) *FavoriteStore {
	return &FavoriteStore{DB: db}
}

// ListProductIDs returns [1, 4, ...] for the given user
func (s *FavoriteStore) ListProductIDs(ctx context.Context, userID int) ([]int, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT product_id
		FROM "Favorites"
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *FavoriteStore) Add(ctx context.Context, userID, productID int) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO "Favorites" (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`, userID, productID)
	return err
}

func (s *FavoriteStore) Remove(ctx context.Context, userID, productID int) error {
	_, err := s.DB.Exec(ctx, `
		DELETE FROM "Favorites"
		WHERE user_id = $1 AND product_id = $2
	`, userID, productID)
	return err
}