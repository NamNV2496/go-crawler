package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/namnv2496/seo/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBTxHandleFunc func(ctx context.Context, tx *gorm.DB) error

type IDatabase interface {
	GetDB() *gorm.DB
	StartTransaction() *gorm.DB
	RunWithTransaction(ctx context.Context, funcs ...DBTxHandleFunc) error
}

type Database struct {
	db *gorm.DB
}

func NewDatabase(
	conf *configs.Config,
) *Database {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		conf.DatabaseConfig.Host,
		conf.DatabaseConfig.User,
		conf.DatabaseConfig.Password,
		conf.DatabaseConfig.DBName,
		conf.DatabaseConfig.Port,
		conf.DatabaseConfig.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}
	return &Database{db: db}
}

func (_self *Database) GetDB() *gorm.DB {
	return _self.db
}

func (_self *Database) StartTransaction() *gorm.DB {
	return _self.db.Begin()
}

func (_self *Database) RunWithTransaction(ctx context.Context, funcs ...DBTxHandleFunc) error {
	timeoutctx, _ := context.WithTimeout(ctx, 5*time.Second)
	tx := _self.db.Begin().WithContext(timeoutctx)
	for _, f := range funcs {
		if err := f(ctx, tx); err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
