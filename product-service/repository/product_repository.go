package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/services/pkg/logger"
)

// Product represents a product entity in the database.
type Product struct {
	ID          string
	Name        string
	Description string
	Price       float64
	ImageURL    string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductRepository handles product database operations.
type ProductRepository struct {
	db  *sql.DB
	log *logger.Logger
}

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *sql.DB, log *logger.Logger) *ProductRepository {
	return &ProductRepository{db: db, log: log}
}

// Migrate creates the products table if it doesn't exist.
func (r *ProductRepository) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL DEFAULT 0,
			image_url TEXT,
			category VARCHAR(100),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to run product migration: %w", err)
	}
	r.log.Info("Product table migration completed")
	return nil
}

// Create inserts a new product.
func (r *ProductRepository) Create(p *Product) (*Product, error) {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	query := `
		INSERT INTO products (id, name, description, price, image_url, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(query,
		p.ID, p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.CreatedAt, p.UpdatedAt,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	r.log.Info("Product created: %s (%s)", p.Name, p.ID)
	return p, nil
}

// GetByID fetches a product by its ID.
func (r *ProductRepository) GetByID(id string) (*Product, error) {
	query := `
		SELECT id, name, description, price, image_url, category, created_at, updated_at
		FROM products WHERE id = $1
	`
	p := &Product{}
	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return p, nil
}

// List fetches products with pagination and optional category filter.
func (r *ProductRepository) List(page, pageSize int, category string) ([]*Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM products"
	listQuery := `
		SELECT id, name, description, price, image_url, category, created_at, updated_at
		FROM products
	`

	if category != "" {
		countQuery += " WHERE category = $1"
		listQuery += " WHERE category = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"
		err := r.db.QueryRow(countQuery, category).Scan(&totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count products: %w", err)
		}
		rows, err := r.db.Query(listQuery, category, pageSize, offset)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to list products: %w", err)
		}
		defer rows.Close()
		return r.scanProducts(rows, totalCount)
	}

	listQuery += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}
	rows, err := r.db.Query(listQuery, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()
	return r.scanProducts(rows, totalCount)
}

func (r *ProductRepository) scanProducts(rows *sql.Rows, totalCount int) ([]*Product, int, error) {
	var products []*Product
	for rows.Next() {
		p := &Product{}
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, totalCount, nil
}

// Update modifies an existing product.
func (r *ProductRepository) Update(p *Product) (*Product, error) {
	p.UpdatedAt = time.Now()
	query := `
		UPDATE products
		SET name = $2, description = $3, price = $4, image_url = $5, category = $6, updated_at = $7
		WHERE id = $1
		RETURNING id, name, description, price, image_url, category, created_at, updated_at
	`
	err := r.db.QueryRow(query,
		p.ID, p.Name, p.Description, p.Price, p.ImageURL, p.Category, p.UpdatedAt,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}
	r.log.Info("Product updated: %s", p.ID)
	return p, nil
}

// Delete removes a product by its ID.
func (r *ProductRepository) Delete(id string) error {
	query := "DELETE FROM products WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("product not found: %s", id)
	}
	r.log.Info("Product deleted: %s", id)
	return nil
}
