package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CartItem struct {
	CartItemID int     `json:"cart_item_id"`
	ProductID  int     `json:"productId"`
	VariantID  int     `json:"variant_id"`
	Name       string  `json:"name"`
	Size       string  `json:"size"`
	Color      string  `json:"color"`
	Price      float64 `json:"price"`
	Image      string  `json:"image"`
	Qty        int     `json:"qty"`
	Key        string  `json:"key"`
}

type OrderDetailItem struct {
	Name      string  `json:"name"`
	ProductID int     `json:"productId"`
	Size      string  `json:"size"`
	Color     string  `json:"color"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

type UserOrder struct {
	ID       string            `json:"id"`
	UserID   int               `json:"userId"`
	Customer string            `json:"customer"`
	Total    float64           `json:"total"`
	Status   string            `json:"status"`
	Date     time.Time         `json:"date"`
	Items    []OrderDetailItem `json:"items"`
}

type OrderStore struct {
	DB *pgxpool.Pool
}

func NewOrderStore(db *pgxpool.Pool) *OrderStore {
	return &OrderStore{DB: db}
}

// ---------- Cart Operations ----------

func (s *OrderStore) GetCart(ctx context.Context, userID int) ([]CartItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.cart_item_id, p.product_id, v.variant_id, p.product_name,
		       COALESCE(v.size, ''), COALESCE(v.color, ''), p.price::float8,
		       COALESCE(p.image_url, ''), c.quantity
		FROM cart_items c
		JOIN product_variants v ON v.variant_id = c.variant_id
		JOIN "Products" p ON p.product_id = v.product_id
		WHERE c.user_id = $1
		ORDER BY c.cart_item_id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []CartItem{}
	for rows.Next() {
		var it CartItem
		if err := rows.Scan(&it.CartItemID, &it.ProductID, &it.VariantID, &it.Name, &it.Size, &it.Color, &it.Price, &it.Image, &it.Qty); err != nil {
			return nil, err
		}
		it.Key = it.Size + "-" + it.Color
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *OrderStore) AddToCart(ctx context.Context, userID, variantID, qty int) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO cart_items (user_id, variant_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, variant_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
	`, userID, variantID, qty)
	return err
}

func (s *OrderStore) UpdateCartItem(ctx context.Context, userID, variantID, qty int) error {
	if qty <= 0 {
		_, err := s.DB.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1 AND variant_id = $2`, userID, variantID)
		return err
	}
	_, err := s.DB.Exec(ctx, `
		UPDATE cart_items
		SET quantity = $1
		WHERE user_id = $2 AND variant_id = $3
	`, qty, userID, variantID)
	return err
}

func (s *OrderStore) ClearCart(ctx context.Context, userID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID)
	return err
}

// ---------- Order & Checkout ----------

type PlaceOrderItem struct {
	VariantID int     `json:"variant_id"`
	ProductID int     `json:"productId"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

func (s *OrderStore) PlaceOrder(ctx context.Context, userID int, address string, items []PlaceOrderItem) (int, error) {
	if len(items) == 0 {
		return 0, errors.New("cart is empty")
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var total float64
	for _, it := range items {
		total += it.Price * float64(it.Qty)
	}

	var orderID int
	err = tx.QueryRow(ctx, `
		INSERT INTO "Orders" (user_id, total_amount, order_status, shipping_address_snapshot)
		VALUES ($1, $2, 'Confirmed', $3)
		RETURNING order_id
	`, userID, total, address).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	for _, it := range items {
		// Insert into Order_Items
		if _, err := tx.Exec(ctx, `
			INSERT INTO "Order_Items" (order_id, variant_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4)
		`, orderID, it.VariantID, it.Qty, it.Price); err != nil {
			return 0, err
		}

		// Deduct stock quantity
		if _, err := tx.Exec(ctx, `
			UPDATE product_variants
			SET stock_quantity = GREATEST(0, stock_quantity - $1)
			WHERE variant_id = $2
		`, it.Qty, it.VariantID); err != nil {
			return 0, err
		}
	}

	// Empty the cart
	if _, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1`, userID); err != nil {
		return 0, err
	}

	return orderID, tx.Commit(ctx)
}

func (s *OrderStore) ListMyOrders(ctx context.Context, userID int) ([]UserOrder, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT o.order_id, u.full_name, o.total_amount::float8, o.order_status, o.order_date
		FROM "Orders" o
		JOIN "User" u ON u.user_id = o.user_id
		WHERE o.user_id = $1
		ORDER BY o.order_date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []UserOrder{}
	orderIDs := []int{}
	orderIndex := map[int]int{}

	for rows.Next() {
		var o UserOrder
		var oid int
		if err := rows.Scan(&oid, &o.Customer, &o.Total, &o.Status, &o.Date); err != nil {
			return nil, err
		}
		o.ID = "YC-" + time.Now().Format("2006") + "-" + string(rune(oid))
		o.UserID = userID
		o.Items = []OrderDetailItem{}
		orderIndex[oid] = len(orders)
		orderIDs = append(orderIDs, oid)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load order items
	for oid, idx := range orderIndex {
		iRows, err := s.DB.Query(ctx, `
			SELECT p.product_name, p.product_id, COALESCE(v.size, ''), COALESCE(v.color, ''),
			       oi.quantity, oi.unit_price::float8
			FROM "Order_Items" oi
			JOIN product_variants v ON v.variant_id = oi.variant_id
			JOIN "Products" p ON p.product_id = v.product_id
			WHERE oi.order_id = $1
		`, oid)
		if err != nil {
			continue
		}
		for iRows.Next() {
			var it OrderDetailItem
			if err := iRows.Scan(&it.Name, &it.ProductID, &it.Size, &it.Color, &it.Qty, &it.Price); err == nil {
				orders[idx].Items = append(orders[idx].Items, it)
			}
		}
		iRows.Close()
	}

	return orders, nil
}