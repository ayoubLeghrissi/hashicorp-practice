package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/services/pkg/logger"
)

// InventoryItem represents an inventory record in the database.
type InventoryItem struct {
	ID                string
	ProductID         string
	Quantity          int32
	WarehouseLocation string
	ReorderLevel      int32
	UpdatedAt         time.Time
}

// InventoryRepository handles inventory database operations.
type InventoryRepository struct {
	db  *sql.DB
	log *logger.Logger
}

// NewInventoryRepository creates a new InventoryRepository.
func NewInventoryRepository(db *sql.DB, log *logger.Logger) *InventoryRepository {
	return &InventoryRepository{db: db, log: log}
}

// Migrate creates the inventory table if it doesn't exist.
func (r *InventoryRepository) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS inventory (
			id UUID PRIMARY KEY,
			product_id VARCHAR(255) NOT NULL UNIQUE,
			quantity INTEGER NOT NULL DEFAULT 0,
			warehouse_location VARCHAR(255),
			reorder_level INTEGER NOT NULL DEFAULT 10,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_inventory_product_id ON inventory(product_id);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run inventory migration: %w", err)
	}
	r.log.Info("Inventory table migration completed")
	return nil
}

// Add inserts a new inventory item.
func (r *InventoryRepository) Add(item *InventoryItem) (*InventoryItem, error) {
	item.ID = uuid.New().String()
	item.UpdatedAt = time.Now()

	query := `
		INSERT INTO inventory (id, product_id, quantity, warehouse_location, reorder_level, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (product_id) DO UPDATE SET
			quantity = inventory.quantity + EXCLUDED.quantity,
			warehouse_location = EXCLUDED.warehouse_location,
			reorder_level = EXCLUDED.reorder_level,
			updated_at = NOW()
		RETURNING id, product_id, quantity, warehouse_location, reorder_level, updated_at
	`
	err := r.db.QueryRow(query,
		item.ID, item.ProductID, item.Quantity, item.WarehouseLocation, item.ReorderLevel, item.UpdatedAt,
	).Scan(&item.ID, &item.ProductID, &item.Quantity, &item.WarehouseLocation, &item.ReorderLevel, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add inventory: %w", err)
	}

	r.log.Info("Inventory added/updated for product %s: qty=%d", item.ProductID, item.Quantity)
	return item, nil
}

// GetByProductID fetches inventory for a specific product.
func (r *InventoryRepository) GetByProductID(productID string) (*InventoryItem, error) {
	query := `
		SELECT id, product_id, quantity, warehouse_location, reorder_level, updated_at
		FROM inventory WHERE product_id = $1
	`
	item := &InventoryItem{}
	err := r.db.QueryRow(query, productID).Scan(
		&item.ID, &item.ProductID, &item.Quantity, &item.WarehouseLocation, &item.ReorderLevel, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inventory not found for product: %s", productID)
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return item, nil
}

// List fetches all inventory items with pagination.
func (r *InventoryRepository) List(page, pageSize int) ([]*InventoryItem, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var totalCount int
	err := r.db.QueryRow("SELECT COUNT(*) FROM inventory").Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count inventory: %w", err)
	}

	query := `
		SELECT id, product_id, quantity, warehouse_location, reorder_level, updated_at
		FROM inventory ORDER BY updated_at DESC LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list inventory: %w", err)
	}
	defer rows.Close()

	var items []*InventoryItem
	for rows.Next() {
		item := &InventoryItem{}
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.WarehouseLocation, &item.ReorderLevel, &item.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan inventory: %w", err)
		}
		items = append(items, item)
	}
	return items, totalCount, nil
}

// UpdateQuantity atomically changes the quantity of an inventory item.
func (r *InventoryRepository) UpdateQuantity(productID string, quantityChange int32) (*InventoryItem, error) {
	query := `
		UPDATE inventory
		SET quantity = quantity + $2, updated_at = NOW()
		WHERE product_id = $1 AND quantity + $2 >= 0
		RETURNING id, product_id, quantity, warehouse_location, reorder_level, updated_at
	`
	item := &InventoryItem{}
	err := r.db.QueryRow(query, productID, quantityChange).Scan(
		&item.ID, &item.ProductID, &item.Quantity, &item.WarehouseLocation, &item.ReorderLevel, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("insufficient inventory or product not found: %s", productID)
		}
		return nil, fmt.Errorf("failed to update inventory quantity: %w", err)
	}

	r.log.Info("Inventory quantity updated for product %s: change=%d, new_qty=%d", productID, quantityChange, item.Quantity)
	return item, nil
}

// Deduct atomically deducts quantity from inventory (for purchases).
func (r *InventoryRepository) Deduct(productID string, quantity int32) (int32, error) {
	query := `
		UPDATE inventory
		SET quantity = quantity - $2, updated_at = NOW()
		WHERE product_id = $1 AND quantity >= $2
		RETURNING quantity
	`
	var remaining int32
	err := r.db.QueryRow(query, productID, quantity).Scan(&remaining)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("insufficient stock for product %s (requested: %d)", productID, quantity)
		}
		return 0, fmt.Errorf("failed to deduct inventory: %w", err)
	}

	r.log.Info("Inventory deducted for product %s: qty=%d, remaining=%d", productID, quantity, remaining)
	return remaining, nil
}

// CheckStock checks if sufficient quantity is available.
func (r *InventoryRepository) CheckStock(productID string, requestedQty int32) (bool, int32, error) {
	item, err := r.GetByProductID(productID)
	if err != nil {
		return false, 0, err
	}
	return item.Quantity >= requestedQty, item.Quantity, nil
}
