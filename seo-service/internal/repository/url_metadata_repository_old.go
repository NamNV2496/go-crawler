package repository

// import (
// 	"context"

// 	"github.com/namnv2496/seo/internal/domain"
// 	"gorm.io/gorm"
// )

// type IUrlMetadataRepository interface {
// 	CreateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error
// 	GetUrlMetadata(ctx context.Context, urlId int64) ([]*domain.UrlMetadata, error)
// 	GetUrlMetadatas(ctx context.Context, urlIds []int64) ([]*domain.UrlMetadata, error)
// 	UpdateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error
// 	DeleteUrlMetadataById(ctx context.Context, tx *gorm.DB, urlId int64) error
// }

// type UrlMetadataRepository struct {
// 	db *gorm.DB
// }

// func NewUrlMetadataRepository(
// 	database IDatabase,
// ) *UrlMetadataRepository {
// 	database.GetDB().AutoMigrate(&domain.UrlMetadata{})
// 	return &UrlMetadataRepository{
// 		db: database.GetDB(),
// 	}
// }

// var _ IUrlMetadataRepository = &UrlMetadataRepository{}

// func (_self *UrlMetadataRepository) CreateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error {
// 	if len(urlMetadata) == 0 {
// 		return nil
// 	}
// 	return tx.Create(&urlMetadata).Error
// }

// func (_self *UrlMetadataRepository) GetUrlMetadata(ctx context.Context, urlId int64) ([]*domain.UrlMetadata, error) {
// 	var resp []*domain.UrlMetadata
// 	err := _self.db.WithContext(ctx).Where("url_id = ?", urlId).Find(&resp).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func (_self *UrlMetadataRepository) GetUrlMetadatas(ctx context.Context, urlIds []int64) ([]*domain.UrlMetadata, error) {
// 	var urls []*domain.UrlMetadata
// 	err := _self.db.WithContext(ctx).Where("url_id IN ?", urlIds).Find(&urls).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return urls, nil
// }

// func (_self *UrlMetadataRepository) UpdateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error {
// 	if len(urlMetadata) == 0 {
// 		return nil
// 	}
// 	// =============== WAY 1: use if setup primary key ===============
// 	// for _, url := range urlMetadata {
// 	// 	err := tx.Clauses(clause.OnConflict{
// 	// 		Where: clause.Where{
// 	// 			Exprs: []clause.Expression{
// 	// 				clause.And(
// 	// 					clause.Eq{
// 	// 						Column: "url_metadata.id",
// 	// 						Value:  url.Id,
// 	// 					},
// 	// 				),
// 	// 			},
// 	// 		},
// 	// 		Columns: []clause.Column{
// 	// 			{Name: "value"},
// 	// 			{Name: "keyword"},
// 	// 		},
// 	// 		DoUpdates: clause.AssignmentColumns([]string{"keyword", "value"}),
// 	// 	}).Create(
// 	// 		domain.UrlMetadata{
// 	// 			UrlId:   url.UrlId,
// 	// 			Keyword: url.Keyword,
// 	// 			Value:   url.Value,
// 	// 		},
// 	// 	).Error
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// }
// 	// =============== WAY 1: manual ===============
// 	inserts, updates, err := _self.getInsertUpdateMetadata(ctx, urlMetadata)
// 	if err != nil {
// 		return err
// 	}
// 	if len(inserts) > 0 {
// 		if err := _self.CreateUrlMetadata(ctx, tx, inserts); err != nil {
// 			return err
// 		}
// 	}
// 	if len(updates) > 0 {
// 		for _, metaData := range updates {
// 			if err := _self.UpdateUrlMetadataById(ctx, tx, metaData); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (_self *UrlMetadataRepository) UpdateUrlMetadataById(ctx context.Context, tx *gorm.DB, urlMetadata *domain.UrlMetadata) error {
// 	return tx.Where("url_id =?", urlMetadata.UrlId).Updates(&urlMetadata).Error
// }

// func (_self *UrlMetadataRepository) DeleteUrlMetadataById(ctx context.Context, tx *gorm.DB, urlId int64) error {
// 	return tx.Where("url_id = ?", urlId).Delete(&domain.Url{}).Error
// }

// func (_self *UrlMetadataRepository) getInsertUpdateMetadata(ctx context.Context, request []*domain.UrlMetadata) ([]*domain.UrlMetadata, []*domain.UrlMetadata, error) {
// 	metaDatas, err := _self.GetUrlMetadata(ctx, request[0].UrlId)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	var insertRequest []*domain.UrlMetadata
// 	var updateRequest []*domain.UrlMetadata
// 	var updateRequestMap = make(map[string]*domain.UrlMetadata)
// 	for _, metaData := range metaDatas {
// 		updateRequestMap[metaData.Keyword] = metaData
// 	}
// 	for _, metaData := range request {
// 		if updateRequestMap[metaData.Keyword] == nil {
// 			insertRequest = append(insertRequest, metaData)
// 		}
// 		oldMetaData := updateRequestMap[metaData.Keyword]
// 		if oldMetaData != nil && (oldMetaData.Value != metaData.Value || oldMetaData.Keyword != metaData.Keyword) {
// 			updateRequest = append(updateRequest, metaData)
// 		}
// 	}
// 	return insertRequest, updateRequest, nil
// }
