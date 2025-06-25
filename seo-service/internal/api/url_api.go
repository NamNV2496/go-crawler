package api

import "github.com/namnv2496/seo/internal/entity"

type CreateUrlRequest struct {
	Url         string                      `query:"url" validate:"required"`
	Name        string                      `query:"name"`
	Tittle      string                      `query:"tittle"`
	Description string                      `query:"description"`
	Template    string                      `query:"template" validate:"required"`
	Prefix      string                      `query:"prefix"`
	Suffix      string                      `query:"suffix"`
	MetaData    []*CreateUrlRequestMetadata `query:"metadata"`
	Domain      string                      `query:"domain"`
	IsActive    bool                        `query:"is_active"`
}

type CreateUrlRequestMetadata struct {
	Keyword string `query:"column:keyword" json:"keyword"`
	Value   string `query:"column:value" json:"value"`
}

type UpdateUrlRequest struct {
	Id          int64                       `param:"id" validate:"required"`
	Url         string                      `query:"url" validate:"required"`
	Name        string                      `query:"name"`
	Tittle      string                      `query:"tittle"`
	Description string                      `query:"description"`
	Template    string                      `query:"template" validate:"required"`
	Prefix      string                      `query:"prefix"`
	Suffix      string                      `query:"suffix"`
	MetaData    []*CreateUrlRequestMetadata `query:"metadata"`
	Domain      string                      `query:"domain"`
	IsActive    bool                        `query:"is_active"`
}

type UpdateUrlRequestMetadata struct {
	Id      int64  `query:"id" json:"id"`
	UrlId   int64  `query:"url_id" json:"url_id"`
	Keyword string `query:"column:keyword" json:"keyword"`
	Value   string `query:"column:value" json:"value"`
}

type UpdateUrlResponse struct {
	Status string `json:"status"`
}

type GetUrlRequest struct {
	Url string `query:"url" validate:"required"`
}

type ParseUrlRequest struct {
	Url string `query:"url" validate:"required"`
}

type ParseUrlResponse struct {
	Uri         string `json:"uri"`
	Path        string `json:"path"`
	Tittle      string `json:"tittle"`
	Description string `json:"description"`
}

type DynamicParamRequest struct {
	Kind     string `query:"kind"`
	Category string `query:"category"`
	City     string `query:"city"`
	Product  string `query:"product"`
	Brand    string `query:"brand"`
	Year     string `query:"year"`
}

type DynamicParamResponse struct {
	Data []*DynamicParamGroup `json:"data"`
}

type DynamicParamGroup struct {
	Group string              `json:"group"`
	Total int                 `json:"total"`
	Data  []*DynamicParamData `json:"data"`
}
type DynamicParamData struct {
	Tittle string `json:"tittle"`
	Uri    string `json:"uri"`
}

type BuildUrlRequest struct {
	Type     string `query:"type"`
	Kind     string `query:"kind"` // validate:"enum=mua-ban,trao-doi,mua,ban,cho-thue"
	City     string `query:"city"`
	Product  string `query:"product"`
	Category string `query:"category"`
	Brand    string `query:"brand"`
	Year     string `query:"year"`
	Month    string `query:"month"`
}

type BuildUrlResponse struct {
	Urls []string `json:"urls"`
}

type GetUrlsRequest struct {
	Page  int `query:"page" validate:"min=1,max=100"`
	Limit int `query:"limit" validate:"min=1,max=100"`
}

type GetUrlsResponse struct {
	Total       int           `json:"total"`
	CurrentPage int           `json:"current_page"`
	Limit       int           `json:"limit"`
	Urls        []*entity.Url `json:"urls"`
}
