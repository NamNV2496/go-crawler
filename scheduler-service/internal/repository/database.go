package repository

import (
	"fmt"

	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type IDatabase interface {
	GetDB() *gorm.DB
	Close() error
}
type Database struct {
	db *gorm.DB
}

func NewDatabase(
	config *configs.Config,
) (*Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		config.DatabaseConfig.Host,
		config.DatabaseConfig.User,
		config.DatabaseConfig.Password,
		config.DatabaseConfig.DBName,
		config.DatabaseConfig.Port,
		config.DatabaseConfig.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		fmt.Printf("Failed to connect to database: %v", err)
		return nil, err
	}
	// Enable connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	db.AutoMigrate(&domain.SchedulerEvent{})
	return &Database{db: db}, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
