package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection
var DB *gorm.DB

// RawDB is the underlying sql.DB connection
var RawDB *sql.DB

// InitDB initializes the database connection
func InitDB() (*sql.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "event_feedback")
	sslMode := getEnv("DB_SSL_MODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode,
	)

	// Connect using GORM
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	DB = gormDB

	// Get the underlying *sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB instance: %w", err)
	}
	RawDB = sqlDB

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Run migrations
	err = runMigrations(gormDB)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return sqlDB, nil
}

// runMigrations creates tables if they don't exist
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&Event{},
		&Form{},
		&FormField{},
		&Submission{},
		&SubmissionResponse{},
	)
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
