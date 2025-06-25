package urlbuilderfactory

import (
	"context"

	"github.com/namnv2496/seo/internal/entity"
	"github.com/namnv2496/seo/internal/repository"
)

type QueryOption struct {
	Field string
	And   bool // true: and, false: or
}

type IBuilder interface {
	Build(ctx context.Context, request map[string]string) ([]*entity.ShortLink, error)
	BuildRecommend(ctx context.Context, request map[string]string, fields []QueryOption) ([]*entity.ShortLink, error)
}

func BuilderFactory(
	kind string,
	repo repository.IShortLinkRepo,
) (IBuilder, error) {
	switch kind {
	case entity.UrlKindCity:
		return &CityBuilder{
			repo: repo,
		}, nil
	case entity.UrlKindProduct:
		return &ProductBuilder{
			repo: repo,
		}, nil
	case entity.UrlKindCategory:
		return &CategoryBuilder{
			repo: repo,
		}, nil
	case entity.UrlKindBrand:
		return &BrandBuilder{
			repo: repo,
		}, nil
	case entity.UrlKindYear:
		return &YearBuilder{
			repo: repo,
		}, nil
	default:
		return nil, nil
	}
}
