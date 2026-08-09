package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/services/pkg/logger"
)

// Cart represents a user's cart.
type Cart struct {
	ID        string
	UserID    string
	Status    string // "active", "checked_out", "abandoned"
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CartItem represents an item in a user's cart.
type CartItem struct {
	ID        string
	CartID    string
	ProductID string
	Quantity  int32
	UnitPrice float64
}

// CartRepository handles cart database operations.
type CartRepository struct {
	db  *sql.DB
	log *logger.Logger
}

// NewCartRepository creates a new CartRepository.
func NewCartRepository(db *sql.DB, log *logger.Logger) *CartRepository {
	return &CartRepository{db: db, log: log}
}

// Migrate creates the cart tables if they don't exist.
func (r *CartRepository) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS carts (
			id UUID PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);
		CREATE INDEX IF NOT EXISTS idx_carts_status ON carts(status);

		CREATE TABLE IF NOT EXISTS cart_items (
			id UUID PRIMARY KEY,
			cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
			product_id VARCHAR(255) NOT NULL,
			quantity INTEGER NOT NULL DEFAULT 1,
			unit_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
			UNIQUE(cart_id, product_id)
		);
		CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run cart migration: %w", err)
	}
	r.log.Info("Cart tables migration completed")
	return nil
}

// GetOrCreateActiveCart gets or creates an active cart for a user.
func (r *CartRepository) GetOrCreateActiveCart(userID string) (*Cart, error) {
	cart := &Cart{}
	err := r.db.QueryRow(
		"SELECT id, user_id, status, created_at, updated_at FROM carts WHERE user_id = $1 AND status = 'active'",
		userID,
	).Scan(&cart.ID, &cart.UserID, &cart.Status, &cart.CreatedAt, &cart.UpdatedAt)

	if err == sql.ErrNoRows {
		cart.ID = uuid.New().String()
		cart.UserID = userID
		cart.Status = "active"
		cart.CreatedAt = time.Now()
		cart.UpdatedAt = time.Now()

		_, err := r.db.Exec(
			"INSERT INTO carts (id, user_id, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			cart.ID, cart.UserID, cart.Status, cart.CreatedAt, cart.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create cart: %w", err)
		}
		r.log.Info("New cart created for user %s: %s", userID, cart.ID)
		return cart, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return cart, nil
}

// GetCartItems fetches all items in a cart.
func (r *CartRepository) GetCartItems(cartID string) ([]*CartItem, error) {
	query := `
		SELECT id, cart_id, product_id, quantity, unit_price
		FROM cart_items WHERE cart_id = $1
	`
	rows, err := r.db.Query(query, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	var items []*CartItem
	for rows.Next() {
		item := &CartItem{}
		if err := rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// AddItem adds or updates an item in the cart.
func (r *CartRepository) AddItem(cartID, productID string, quantity int32, unitPrice float64) (*CartItem, error) {
	itemID := uuid.New().String()
	query := `
		INSERT INTO cart_items (id, cart_id, product_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cart_id, product_id) DO UPDATE SET
			quantity = cart_items.quantity + EXCLUDED.quantity,
			unit_price = EXCLUDED.unit_price
		RETURNING id, cart_id, product_id, quantity, unit_price
	`
	item := &CartItem{}
	err := r.db.QueryRow(query, itemID, cartID, productID, quantity, unitPrice).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.UnitPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	// Update cart timestamp
	r.db.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)

	r.log.Info("Item added to cart %s: product=%s, qty=%d", cartID, productID, quantity)
	return item, nil
}

// RemoveItem removes an item from the cart.
func (r *CartRepository) RemoveItem(cartID, productID string) error {
	result, err := r.db.Exec(
		"DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2",
		cartID, productID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove item from cart: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found in cart")
	}

	r.db.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	r.log.Info("Item removed from cart %s: product=%s", cartID, productID)
	return nil
}

// UpdateItemQuantity sets the quantity of a cart item.
func (r *CartRepository) UpdateItemQuantity(cartID, productID string, quantity int32) (*CartItem, error) {
	if quantity <= 0 {
		err := r.RemoveItem(cartID, productID)
		return nil, err
	}

	query := `
		UPDATE cart_items SET quantity = $3
		WHERE cart_id = $1 AND product_id = $2
		RETURNING id, cart_id, product_id, quantity, unit_price
	`
	item := &CartItem{}
	err := r.db.QueryRow(query, cartID, productID, quantity).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.UnitPrice,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found in cart")
		}
		return nil, fmt.Errorf("failed to update item quantity: %w", err)
	}

	r.db.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	return item, nil
}

// CheckoutCart marks a cart as checked out and returns a new order ID.
func (r *CartRepository) CheckoutCart(cartID string) (string, error) {
	orderID := uuid.New().String()
	result, err := r.db.Exec(
		"UPDATE carts SET status = 'checked_out', updated_at = NOW() WHERE id = $1 AND status = 'active'",
		cartID,
	)
	if err != nil {
		return "", fmt.Errorf("failed to checkout cart: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "", fmt.Errorf("cart not found or already checked out")
	}

	r.log.Info("Cart checked out: %s, order: %s", cartID, orderID)
	return orderID, nil
}

// ClearCart removes all items from a cart.
func (r *CartRepository) ClearCart(cartID string) error {
	_, err := r.db.Exec("DELETE FROM cart_items WHERE cart_id = $1", cartID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	r.db.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	r.log.Info("Cart cleared: %s", cartID)
	return nil
}

// CalculateTotal computes the total price of a cart.
func (r *CartRepository) CalculateTotal(cartID string) (float64, error) {
	var total sql.NullFloat64
	err := r.db.QueryRow(
		"SELECT SUM(quantity * unit_price) FROM cart_items WHERE cart_id = $1",
		cartID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total: %w", err)
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Float64, nil
}
