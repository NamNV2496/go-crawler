package urlbuilderfactory

import (
	"context"

	"github.com/namnv2496/seo/internal/domain"
	"github.com/namnv2496/seo/internal/entity"
	"github.com/namnv2496/seo/internal/repository"
	"github.com/namnv2496/seo/pkg/utils"
)

type BrandBuilder struct {
	repo repository.IShortLinkRepo
}

func NewBrandBuilder(
	repo repository.IShortLinkRepo,
) *BrandBuilder {
	return &BrandBuilder{
		repo: repo,
	}
}

var _ IBuilder = &BrandBuilder{}

func (_self *BrandBuilder) Build(ctx context.Context, request map[string]string) ([]*entity.ShortLink, error) {
	var opts []repository.QueryOptionFunc
	opts = append(opts, repository.WithCondition("filter ->> 'brand' = ?", request["brand"]))
	opts = append(opts, repository.WithOffset(0))
	opts = append(opts, repository.WithLimit(5))

	result, err := _self.repo.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	var resp []*entity.ShortLink
	for _, shortLink := range result {
		var elem *entity.ShortLink
		utils.Copy(&elem, shortLink)
		resp = append(resp, elem)
	}
	return resp, nil
}

func (_self *BrandBuilder) BuildRecommend(ctx context.Context, request map[string]string, fields []QueryOption) ([]*entity.ShortLink, error) {
	var resp []*entity.ShortLink
	var data []*domain.ShortLink
	// find the same brand name
	brand := request["brand"]
	if brand == "" {
		return nil, nil
	}
	var opts []repository.QueryOptionFunc
	opts = append(opts, repository.WithOffset(0))
	opts = append(opts, repository.WithLimit(5))
	for _, field := range fields {
		if field.And {
			opts = append(opts, repository.WithCondition("filter->>'"+field.Field+"' =?", request[field.Field]))
		} else {
			opts = append(opts, repository.WithOrCondition("filter->>'"+field.Field+"' =?", request[field.Field]))
		}
	}
	data, err := _self.repo.Finds(ctx, opts...)
	if err != nil {
		return nil, err
	}
	for _, shortLink := range data {
		var elem *entity.ShortLink
		utils.Copy(&elem, shortLink)
		resp = append(resp, elem)
	}
	return resp, nil
}
