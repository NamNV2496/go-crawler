package repository

// import (
// 	"context"

// 	"github.com/namnv2496/seo/internal/domain"
// 	"gorm.io/gorm"
// )

// type IShortLinkRepository interface {
// 	GetShortLinks(ctx context.Context, offset, limit int, request map[string]string) ([]*domain.ShortLink, error)
// }

// type ShortLinkRepository struct {
// 	db *gorm.DB
// }

// func NewShortLinkRepository(
// 	database IDatabase,
// ) *ShortLinkRepository {
// 	database.GetDB().AutoMigrate(&domain.ShortLink{})
// 	return &ShortLinkRepository{
// 		db: database.GetDB(),
// 	}
// }

// var _ IShortLinkRepository = &ShortLinkRepository{}

// func (_self *ShortLinkRepository) GetShortLinks(ctx context.Context, offset, limit int, request map[string]string) ([]*domain.ShortLink, error) {
// 	var urlData []*domain.ShortLink
// 	tx := _self.db.WithContext(ctx)
// 	// dynamic filter
// 	for field, value := range request {
// 		tx = tx.Where(field, value)
// 	}
// 	var err error
// 	if offset > 0 && limit > 0 {
// 		err = tx.Offset(offset).Limit(limit).Find(&urlData).Error
// 	} else {
// 		err = tx.Find(&urlData).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return urlData, nil
// }
