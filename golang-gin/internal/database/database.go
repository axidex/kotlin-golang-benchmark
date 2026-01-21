package database

import (
	"log"
	"os"
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

	// Configure GORM logger
	logConfig := gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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
	log.Printf("Connection pool: max open=%d, max idle=%d, max lifetime=%s", 100, 10, time.Hour)
}
