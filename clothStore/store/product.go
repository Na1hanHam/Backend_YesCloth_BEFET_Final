package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductVariant struct {
	VariantID int    `json:"variant_id"`
	ProductID int    `json:"product_id"`
	Size      string `json:"size"`
	Color     string `json:"color"`
	Stock     int    `json:"stock_quantity"`
}

type Product struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	Category    string           `json:"category"`
	Price       float64          `json:"price"`
	Image       string           `json:"image"`
	Description string           `json:"description"`
	IsNew       bool             `json:"isNew"`
	Colors      []string         `json:"colors"`
	Sizes       []string         `json:"sizes"`
	Stock       int              `json:"stock"`
	Variants    []ProductVariant `json:"variants"`
}

type UpsertProductInput struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	Colors      []string `json:"colors"`
	Sizes       []string `json:"sizes"`
	IsNew       bool     `json:"isNew"`
}

type ProductStore struct {
	DB *pgxpool.Pool
}

func NewProductStore(db *pgxpool.Pool) *ProductStore {
	return &ProductStore{DB: db}
}

func (s *ProductStore) List(ctx context.Context) ([]Product, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT p.product_id, p.product_name, COALESCE(c.category_name, ''),
		       p.price::float8, COALESCE(p.image_url, ''), COALESCE(p.description, ''),
		       COALESCE(p.is_new, false)
		FROM "Products" p
		LEFT JOIN "Categories" c ON c.category_id = p.category_id
		ORDER BY p.product_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []Product{}
	indexByID := map[int]int{}

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.Image, &p.Description, &p.IsNew); err != nil {
			return nil, err
		}
		p.Colors = []string{}
		p.Sizes = []string{}
		p.Variants = []ProductVariant{}
		indexByID[p.ID] = len(products)
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load variants and aggregate colors, sizes, and total stock per product
	vRows, err := s.DB.Query(ctx, `
		SELECT variant_id, product_id, COALESCE(size, ''), COALESCE(color, ''), stock_quantity
		FROM product_variants
		ORDER BY variant_id
	`)
	if err != nil {
		return nil, err
	}
	defer vRows.Close()

	for vRows.Next() {
		var v ProductVariant
		if err := vRows.Scan(&v.VariantID, &v.ProductID, &v.Size, &v.Color, &v.Stock); err != nil {
			return nil, err
		}
		idx, ok := indexByID[v.ProductID]
		if !ok {
			continue
		}
		p := &products[idx]
		p.Variants = append(p.Variants, v)
		p.Stock += v.Stock
		if v.Color != "" && !containsStr(p.Colors, v.Color) {
			p.Colors = append(p.Colors, v.Color)
		}
		if v.Size != "" && !containsStr(p.Sizes, v.Size) {
			p.Sizes = append(p.Sizes, v.Size)
		}
	}
	return products, vRows.Err()
}

func containsStr(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (s *ProductStore) Create(ctx context.Context, in UpsertProductInput) (int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var catID *int
	_ = tx.QueryRow(ctx, `SELECT category_id FROM "Categories" WHERE category_name = $1 LIMIT 1`, in.Category).Scan(&catID)
	var productID int
	err = tx.QueryRow(ctx, `
		INSERT INTO "Products" (category_id, product_name, price, image_url, description, is_new)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING product_id
	`, catID, in.Name, in.Price, in.Image, in.Description, in.IsNew).Scan(&productID)
	if err != nil {
		return 0, err
	}
	colors := in.Colors
	if len(colors) == 0 {
		colors = []string{"White", "Black"}
	}
	sizes := in.Sizes
	if len(sizes) == 0 {
		sizes = []string{"S", "M", "L"}
	}
	totalCombos := len(colors) * len(sizes)
	baseStock := in.Stock / totalCombos
	rem := in.Stock % totalCombos
	idx := 0
	for _, c := range colors {
		for _, sz := range sizes {
			st := baseStock
			if idx < rem {
				st++
			}
			idx++
			if _, err := tx.Exec(ctx, `
				INSERT INTO product_variants (product_id, color, size, stock_quantity)
				VALUES ($1, $2, $3, $4)
			`, productID, c, sz, st); err != nil {
				return 0, err
			}
		}
	}
	return productID, tx.Commit(ctx)
}
func (s *ProductStore) Update(ctx context.Context, productID int, in UpsertProductInput) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var catID *int
	_ = tx.QueryRow(ctx, `SELECT category_id FROM "Categories" WHERE category_name = $1 LIMIT 1`, in.Category).Scan(&catID)
	_, err = tx.Exec(ctx, `
		UPDATE "Products"
		SET category_id = $1, product_name = $2, price = $3, image_url = $4, description = $5
		WHERE product_id = $6
	`, catID, in.Name, in.Price, in.Image, in.Description, productID)
	if err != nil {
		return err
	}
	// Distribute updated total stock across the product's variants
	var count int
	_ = tx.QueryRow(ctx, `SELECT COUNT(*) FROM product_variants WHERE product_id = $1`, productID).Scan(&count)
	if count > 0 {
		perVariant := in.Stock / count
		_, err = tx.Exec(ctx, `UPDATE product_variants SET stock_quantity = $1 WHERE product_id = $2`, perVariant, productID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (s *ProductStore) Delete(ctx context.Context, productID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM "Products" WHERE product_id = $1`, productID)
	return err
}