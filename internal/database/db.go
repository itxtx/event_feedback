package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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

	// First connect to PostgreSQL server without specifying a database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, sslMode,
	)

	// Connect to PostgreSQL server
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Now connect to the specific database with proper DSN format
	dsn = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, sslMode,
	)

	// Connect using GORM with proper configuration
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
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
	// Drop existing tables in reverse order of dependencies
	tables := []string{
		"submission_responses",
		"submissions",
		"form_fields",
		"forms",
		"events",
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Auto migrate all models
	err := db.AutoMigrate(
		&Event{},
		&Form{},
		&FormField{},
		&Submission{},
		&SubmissionResponse{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed test data
	err = seedTestData(db)
	if err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}

	return nil
}

// seedTestData adds some test data to the database
func seedTestData(db *gorm.DB) error {
	// Create a test event
	event := Event{
		Name:        "Test Event",
		Description: "This is a test event for the feedback system",
		Date:        time.Now().AddDate(0, 1, 0), // One month from now
	}
	result := db.Create(&event)
	if result.Error != nil {
		return fmt.Errorf("failed to create test event: %w", result.Error)
	}

	// Create a test form
	form := Form{
		EventID:     event.ID,
		Title:       "Test Feedback Form",
		IsMultiStep: false,
		IsPublished: true,
	}
	result = db.Create(&form)
	if result.Error != nil {
		return fmt.Errorf("failed to create test form: %w", result.Error)
	}

	// Create some test form fields
	fields := []FormField{
		{
			FormID:     form.ID,
			FieldType:  "text",
			Label:      "What did you like most about the event?",
			IsRequired: true,
			FieldOrder: 1,
		},
		{
			FormID:     form.ID,
			FieldType:  "textarea",
			Label:      "Do you have any suggestions for improvement?",
			IsRequired: false,
			FieldOrder: 2,
		},
		{
			FormID:     form.ID,
			FieldType:  "select",
			Label:      "How would you rate the event?",
			Options:    `["Excellent","Good","Average","Poor"]`,
			IsRequired: true,
			FieldOrder: 3,
		},
	}

	for _, field := range fields {
		result = db.Create(&field)
		if result.Error != nil {
			return fmt.Errorf("failed to create test form field: %w", result.Error)
		}
	}

	return nil
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
