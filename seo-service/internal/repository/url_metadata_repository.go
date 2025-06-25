package repository

import (
	"context"

	"github.com/namnv2496/seo/internal/domain"
	"gorm.io/gorm"
)

type IUrlMetadataRepo interface {
	IRepository[domain.UrlMetadata]
	CreateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error
	GetUrlMetadata(ctx context.Context, urlId int64) ([]*domain.UrlMetadata, error)
	GetUrlMetadatas(ctx context.Context, urlIds []int64) ([]*domain.UrlMetadata, error)
	UpdateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error
	DeleteUrlMetadataById(ctx context.Context, tx *gorm.DB, urlId int64) error
}

type UrlMetadataRepo struct {
	baseRepository[domain.UrlMetadata]
}

func NewUrlMetadataRepo(
	database IDatabase,
) *UrlMetadataRepo {
	database.GetDB().AutoMigrate(&domain.UrlMetadata{})
	return &UrlMetadataRepo{
		baseRepository: baseRepository[domain.UrlMetadata]{
			db: database.GetDB(),
		},
	}
}

var _ IUrlMetadataRepo = &UrlMetadataRepo{}

func (_self *UrlMetadataRepo) CreateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error {
	if len(urlMetadata) == 0 {
		return nil
	}
	return _self.Inserts(ctx, urlMetadata)
}

func (_self *UrlMetadataRepo) GetUrlMetadata(ctx context.Context, urlId int64) ([]*domain.UrlMetadata, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("url_id = ?", urlId))
	opts = append(opts, WithOrderBy("id DESC"))
	return _self.Finds(ctx, opts...)
}

func (_self *UrlMetadataRepo) GetUrlMetadatas(ctx context.Context, urlIds []int64) ([]*domain.UrlMetadata, error) {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("url_id IN ?", urlIds))
	opts = append(opts, WithOrderBy("id DESC"))
	return _self.Finds(ctx, opts...)
}

func (_self *UrlMetadataRepo) UpdateUrlMetadata(ctx context.Context, tx *gorm.DB, urlMetadata []*domain.UrlMetadata) error {
	if len(urlMetadata) == 0 {
		return nil
	}
	// =============== WAY 1: use if setup primary key ===============
	// for _, url := range urlMetadata {
	// 	err := tx.Clauses(clause.OnConflict{
	// 		Where: clause.Where{
	// 			Exprs: []clause.Expression{
	// 				clause.And(
	// 					clause.Eq{
	// 						Column: "url_metadata.id",
	// 						Value:  url.Id,
	// 					},
	// 				),
	// 			},
	// 		},
	// 		Columns: []clause.Column{
	// 			{Name: "value"},
	// 			{Name: "keyword"},
	// 		},
	// 		DoUpdates: clause.AssignmentColumns([]string{"keyword", "value"}),
	// 	}).Create(
	// 		domain.UrlMetadata{
	// 			UrlId:   url.UrlId,
	// 			Keyword: url.Keyword,
	// 			Value:   url.Value,
	// 		},
	// 	).Error
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	// =============== WAY 1: manual ===============
	inserts, updates, err := _self.getInsertUpdateMetadata(ctx, urlMetadata)
	if err != nil {
		return err
	}
	if len(inserts) > 0 {
		if err := _self.CreateUrlMetadata(ctx, tx, inserts); err != nil {
			return err
		}
	}
	if len(updates) > 0 {
		for _, metaData := range updates {
			if err := _self.UpdateUrlMetadataById(ctx, tx, metaData); err != nil {
				return err
			}
		}
	}
	return nil
}

func (_self *UrlMetadataRepo) UpdateUrlMetadataById(ctx context.Context, tx *gorm.DB, urlMetadata *domain.UrlMetadata) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("url_id =?", urlMetadata.Id))
	return _self.UpdateOnce(ctx, *urlMetadata, opts...)
}

func (_self *UrlMetadataRepo) DeleteUrlMetadataById(ctx context.Context, tx *gorm.DB, urlId int64) error {
	var opts []QueryOptionFunc
	opts = append(opts, WithCondition("url_id =?", urlId))
	return _self.DeleteById(ctx, domain.UrlMetadata{}, opts...)
}

func (_self *UrlMetadataRepo) getInsertUpdateMetadata(ctx context.Context, request []*domain.UrlMetadata) ([]*domain.UrlMetadata, []*domain.UrlMetadata, error) {
	metaDatas, err := _self.GetUrlMetadata(ctx, request[0].UrlId)
	if err != nil {
		return nil, nil, err
	}
	var insertRequest []*domain.UrlMetadata
	var updateRequest []*domain.UrlMetadata
	var updateRequestMap = make(map[string]*domain.UrlMetadata)
	for _, metaData := range metaDatas {
		updateRequestMap[metaData.Keyword] = metaData
	}
	for _, metaData := range request {
		if updateRequestMap[metaData.Keyword] == nil {
			insertRequest = append(insertRequest, metaData)
		}
		oldMetaData := updateRequestMap[metaData.Keyword]
		if oldMetaData != nil && (oldMetaData.Value != metaData.Value || oldMetaData.Keyword != metaData.Keyword) {
			updateRequest = append(updateRequest, metaData)
		}
	}
	return insertRequest, updateRequest, nil
}
