package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/services/pkg/logger"
)

// Config holds database connection settings.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns a default PostgreSQL config.
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "services",
		SSLMode:  "disable",
	}
}

// Connect establishes a PostgreSQL connection with retry logic for scalability.
func Connect(cfg Config, log *logger.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var db *sql.DB
	var err error

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Warn("Failed to open DB connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Warn("Failed to ping DB (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Connection pool settings for scalability
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	log.Info("Connected to PostgreSQL database: %s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)
	return db, nil
}
