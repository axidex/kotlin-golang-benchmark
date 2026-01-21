package database

import (
	"log"
	"os"
	"strconv"
	"time"

	"dev.sourcecraft.dolgintsev/golang-gin/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormprometheus "gorm.io/plugin/prometheus"
)

var DB *gorm.DB

func Connect() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Default connection for dev
		dsn = "postgres://postgres:postgres@localhost:5432/benchmark?sslmode=disable"
	}

	// Configure GORM logger - only show errors, not INFO/WARN
	logConfig := gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}

	pgConfig := postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}

	DB, err = gorm.Open(postgres.New(pgConfig), &logConfig)

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Get connection pool parameters from environment variables with defaults
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 300)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 50)
	connMaxLifetime := getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour)

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	log.Println("Database connected successfully")

	// Initialize Prometheus plugin for connection pool metrics
	err = DB.Use(gormprometheus.New(gormprometheus.Config{
		DBName:          "benchmark",
		RefreshInterval: 15, // Refresh metrics every 15 seconds
		MetricsCollector: []gormprometheus.MetricsCollector{
			&gormprometheus.Postgres{
				VariableNames: []string{"max_connections"},
			},
		},
	}))
	if err != nil {
		log.Printf("Warning: Failed to initialize Prometheus plugin: %v", err)
	} else {
		log.Println("Prometheus metrics enabled for database connection pool")
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(&models.Product{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")
	log.Printf("Connection pool: max open=%d, max idle=%d, max lifetime=%s", maxOpenConns, maxIdleConns, connMaxLifetime)
}

// Helper function to get environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}

// Helper function to get environment variable as duration with default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid duration for %s, using default %s", key, defaultValue)
		return defaultValue
	}
	return value
}
