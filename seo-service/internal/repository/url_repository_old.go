package repository

// import (
// 	"context"

// 	"github.com/namnv2496/seo/internal/domain"
// 	"gorm.io/gorm"
// )

// type IUrlRepository interface {
// 	CreateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) (int64, error)
// 	GetUrl(ctx context.Context, url string) (*domain.Url, error)
// 	GetUrls(ctx context.Context, offset, limit int) ([]*domain.Url, error)
// 	UpdateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) error
// 	DeleteUrl(ctx context.Context, tx *gorm.DB, url string) error
// }

// type UrlRepository struct {
// 	db *gorm.DB
// }

// func NewUrlRepository(
// 	database IDatabase,
// ) *UrlRepository {
// 	database.GetDB().AutoMigrate(&domain.Url{})
// 	return &UrlRepository{
// 		db: database.GetDB(),
// 	}
// }

// var _ IUrlRepository = &UrlRepository{}

// func (_self *UrlRepository) CreateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) (int64, error) {
// 	tx.Create(&url)
// 	return url.Id, nil
// }

// func (_self *UrlRepository) GetUrl(ctx context.Context, url string) (*domain.Url, error) {
// 	var resp *domain.Url
// 	err := _self.db.WithContext(ctx).Where("url = ? AND is_active = true", url).Find(&resp).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func (_self *UrlRepository) GetUrls(ctx context.Context, offset, limit int) ([]*domain.Url, error) {
// 	var urlData []*domain.Url
// 	var err error
// 	if offset > 0 && limit > 0 {
// 		err = _self.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&urlData).Error
// 	} else {
// 		err = _self.db.WithContext(ctx).Find(&urlData).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	return urlData, nil
// }

// func (_self *UrlRepository) UpdateUrl(ctx context.Context, tx *gorm.DB, url domain.Url) error {
// 	return tx.Model(&url).Updates(url).Error
// }

// func (_self *UrlRepository) DeleteUrl(ctx context.Context, tx *gorm.DB, url string) error {
// 	return tx.Where("url = ?", url).Delete(&domain.Url{}).Error
// }
