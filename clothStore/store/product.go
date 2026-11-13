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