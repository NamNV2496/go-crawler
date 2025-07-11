package repository

import (
	"context"

	"gorm.io/gorm"
)

type IEntity interface {
	TableName() string
}

type LimitOption struct {
	Limit   int
	Offset  int
	OrderBy string
}

type QueryOptionFunc func(tx *gorm.DB) *gorm.DB

type IRepository[E IEntity] interface {
	InsertOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error
	Inserts(ctx context.Context, entities []*E, opts ...QueryOptionFunc) error
	UpdateOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error
	Updates(ctx context.Context, entities []*E, opts ...QueryOptionFunc) error
	DeleteOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error
	DeleteById(ctx context.Context, entity *E, opts ...QueryOptionFunc) error
	Finds(ctx context.Context, opts ...QueryOptionFunc) ([]*E, error)
	Find(ctx context.Context, opts ...QueryOptionFunc) (*E, error)
	CountOnce(ctx context.Context, opts ...QueryOptionFunc) (int64, error)
}

type baseRepository[E IEntity] struct {
	db *gorm.DB
}

func newBaseRepository[E IEntity](db *gorm.DB) baseRepository[E] {
	return baseRepository[E]{
		db: db.Model(new(E)),
	}
}

func (_self *baseRepository[E]) GetDB() *gorm.DB {
	return _self.db
}

func (_self *baseRepository[E]) InsertOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Create(entity).Error
	return err
}

func (_self *baseRepository[E]) Inserts(ctx context.Context, entities []*E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Create(entities).Error
	return err
}

func (_self *baseRepository[E]) UpdateOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Save(entity).Error
	return err
}

func (_self *baseRepository[E]) Updates(ctx context.Context, entities []*E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Save(entities).Error
	return err
}

func (_self *baseRepository[E]) DeleteOnce(ctx context.Context, entity *E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Delete(entity).Error
	return err
}

func (_self *baseRepository[E]) DeleteById(ctx context.Context, entity *E, opts ...QueryOptionFunc) error {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	err := tx.Delete(entity).Error
	return err
}

func (_self *baseRepository[E]) Find(ctx context.Context, opts ...QueryOptionFunc) (*E, error) {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	var entityData *E
	err := tx.First(&entityData).Error
	if err != nil {
		return nil, err
	}
	return entityData, nil
}

func (_self *baseRepository[E]) Finds(ctx context.Context, opts ...QueryOptionFunc) ([]*E, error) {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	var entities []*E
	err := tx.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (_self *baseRepository[E]) CountOnce(ctx context.Context, opts ...QueryOptionFunc) (int64, error) {
	tx := _self.db.WithContext(ctx)
	for _, opt := range opts {
		tx = opt(tx)
	}
	var count int64
	err := tx.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func WithOrderBy(orderBy string) QueryOptionFunc {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Order(orderBy)
	}
}

func WithLimit(limit int) QueryOptionFunc {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Limit(limit)
	}
}

func WithOffset(offset int) QueryOptionFunc {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Offset(offset)
	}
}

func WithCondition(condition string, args ...any) QueryOptionFunc {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(condition, args...)
	}
}

func WithOrCondition(condition string, args ...any) QueryOptionFunc {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Or(condition, args...)
	}
}
