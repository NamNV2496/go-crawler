package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/namnv2496/seo/configs"
	"github.com/namnv2496/seo/internal/domain"
	"github.com/namnv2496/seo/internal/entity"
	"github.com/namnv2496/seo/internal/repository"
	"github.com/namnv2496/seo/internal/service/urlbuilderfactory"
	"github.com/namnv2496/seo/pkg/utils"
	"gorm.io/gorm"
)

type IUrlService interface {
	ParseUrl(ctx context.Context, url string) (*entity.Url, error)
	BuildUrl(ctx context.Context, kind string, request map[string]string) ([]string, error)
	DynamicRecommendParseByUrl(ctx context.Context, request map[string]string) (*entity.DynamicRecommend, error)

	CreateUrl(ctx context.Context, url entity.Url) error
	UpdateUrl(ctx context.Context, url entity.Url) error
	DeleteUrl(ctx context.Context, url entity.Url) error
	GetUrl(ctx context.Context, url string) (*entity.Url, error)
	GetUrls(ctx context.Context, offset, limit int) ([]*entity.Url, error)
}

type UrlService struct {
	aiHost          string
	db              repository.IDatabase
	urlRepo         repository.IUrlRepo
	urlMetadataRepo repository.IUrlMetadataRepo
	shortlinkRepo   repository.IShortLinkRepo
}

func NewUrlService(
	conf *configs.Config,
	db repository.IDatabase,
	urlRepo repository.IUrlRepo,
	urlMetadataRepo repository.IUrlMetadataRepo,
	shortlinkRepo repository.IShortLinkRepo,
) *UrlService {
	return &UrlService{
		aiHost:          conf.AIConfig.Host,
		db:              db,
		urlRepo:         urlRepo,
		urlMetadataRepo: urlMetadataRepo,
		shortlinkRepo:   shortlinkRepo,
	}
}

var _ IUrlService = &UrlService{}

func (_self *UrlService) ParseUrl(ctx context.Context, url string) (*entity.Url, error) {
	urlData, err := _self.urlRepo.GetUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	if urlData == nil {
		return nil, nil
	}
	metadata, err := _self.urlMetadataRepo.GetUrlMetadata(ctx, urlData.Id)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string, 0)
	resp := &entity.Url{}
	utils.Copy(&resp, urlData)
	if metadata != nil {
		var metadataEntity []*entity.UrlMetadata
		for _, meta := range metadata {
			var elem *entity.UrlMetadata
			utils.Copy(&elem, meta)
			metadataEntity = append(metadataEntity, elem)
			params[meta.Keyword] = meta.Value
		}
		resp.Metadata = metadataEntity
	}
	resp.Tittle, _ = utils.BuildByTemplate(ctx, "build-tittle", urlData.Tittle, params)
	resp.Description, _ = utils.BuildByTemplate(ctx, "build-description", urlData.Description, params)
	return resp, nil
}

func (_self *UrlService) BuildUrl(ctx context.Context, kind string, request map[string]string) ([]string, error) {
	switch request["type"] {
	case "ai":
		// TBU: ==================== Intergrate AI to build ====================
		return _self.buildUrlByAI(ctx, kind, request)
	case "template":
		return _self.buildUrlByTemplate(ctx, kind, request)
	case "regex":
		return buildUrlByRegex(ctx, kind, request)
	default:
		return []string{"not-found"}, nil
	}
}

func (_self *UrlService) buildUrlByAI(ctx context.Context, kind string, request map[string]string) ([]string, error) {
	data := ""
	for key, value := range request {
		data += key + ": " + value + "; "
	}

	// Prepare messages
	messages := []entity.Message{
		{
			Role:    "user",
			Content: "You are master SEO",
		},
		{
			Role:    "user",
			Content: "return URL only example: mua-ban-dien-thoai-hn",
		},
		{
			Role:    "user",
			Content: "You must return 3 highest score of url for SEO with my input data: " + data,
		},
	}

	req := entity.AiSEORequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   500,
	}

	requestJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Correct content-type and URL usage
	httpReq, err := http.NewRequestWithContext(ctx, "POST", _self.aiHost, bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(body))
	}

	// 3. Read the response (for now, just print or parse as needed)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result entity.AiSEOResponse
	json.Unmarshal(respBody, &result)
	fmt.Println("AI response:", result.Response)

	return []string{result.Response}, nil
}

func (_self *UrlService) buildUrlByTemplate(ctx context.Context, kind string, request map[string]string) ([]string, error) {
	// Construct the URL template
	urlData, err := _self.urlRepo.GetUrl(ctx, kind)
	if err != nil {
		return []string{}, err
	}
	if urlData == nil {
		return []string{}, nil
	}
	urlTemplate := template.New("url-template")
	urlText := urlData.Template
	if urlData.Prefix != "" {
		urlText = urlData.Prefix + urlData.Template
	}
	if urlData.Suffix != "" {
		urlText = urlText + "-" + urlData.Suffix
	}
	urlTemplate, err = urlTemplate.Parse(urlText)
	if err != nil {
		return []string{}, err
	}
	result := new(bytes.Buffer)
	request["kind"] = kind
	if err := urlTemplate.Execute(result, request); err != nil {
		return []string{}, err
	}
	return []string{result.String()}, nil
}

func buildUrlByRegex(ctx context.Context, kind string, request map[string]string) ([]string, error) {
	var templates = []string{
		"{kind}-{category}-{product}-{brand}-{city}",
		"{kind}-{category}-{product}",
		"{kind}-{product}-{city}",
		"{kind}-{category}-{city}",
		"{kind}-{category}-{brand}-{product}",
		"{kind}-{category}-{year}-{month}",
		"{kind}-{product}",
	}
	request["kind"] = kind
	var urls []string
	for _, tpl := range templates {
		requiredFields := extractor(tpl)
		skip := false
		for _, field := range requiredFields {
			if val, ok := request[field]; !ok || val == "" {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		url := tpl
		for key, val := range request {
			url = strings.ReplaceAll(url, "{"+key+"}", val)
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func extractor(template string) []string {
	re := regexp.MustCompile(`\{(\w+)\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	var fields []string
	for _, match := range matches {
		fields = append(fields, match[1])
	}
	return fields
}

func (_self *UrlService) DynamicRecommendParseByUrl(ctx context.Context, request map[string]string) (*entity.DynamicRecommend, error) {
	dataGroup := make([]*entity.DynamicRecommendGroup, 0)
	// city
	cityBuilder, err := urlbuilderfactory.BuilderFactory(entity.UrlKindCity, _self.shortlinkRepo)
	if err == nil {
		cities, recoErr := cityBuilder.BuildRecommend(ctx, request, []urlbuilderfactory.QueryOption{
			{
				Field: "city",
				And:   true,
			},
			{
				Field: "category",
				And:   true,
			},
		})
		if recoErr == nil && len(cities) > 0 {
			citiesGroup := &entity.DynamicRecommendGroup{
				Group: "Gợi ý cùng thành phố",
				Data:  cities,
			}
			dataGroup = append(dataGroup, citiesGroup)
		}
	}
	// product
	productBuilder, err := urlbuilderfactory.BuilderFactory(entity.UrlKindProduct, _self.shortlinkRepo)
	if err == nil {
		products, recoErr := productBuilder.BuildRecommend(ctx, request, []urlbuilderfactory.QueryOption{
			{
				Field: "product",
				And:   true,
			},
		})
		if recoErr == nil && len(products) > 0 {
			productGroup := &entity.DynamicRecommendGroup{
				Group: "Gợi ý cùng sản phẩm",
				Data:  products,
			}
			dataGroup = append(dataGroup, productGroup)
		}
	}
	// category
	categoryBuilder, err := urlbuilderfactory.BuilderFactory(entity.UrlKindCategory, _self.shortlinkRepo)
	if err == nil {
		categories, recoErr := categoryBuilder.BuildRecommend(ctx, request, []urlbuilderfactory.QueryOption{
			{
				Field: "category",
				And:   true,
			},
		})
		if recoErr == nil && len(categories) > 0 {
			categoryGroup := &entity.DynamicRecommendGroup{
				Group: "Gợi ý cùng danh mục",
				Data:  categories,
			}
			dataGroup = append(dataGroup, categoryGroup)
		}
	}
	// brand
	brandBuilder, err := urlbuilderfactory.BuilderFactory(entity.UrlKindBrand, _self.shortlinkRepo)
	if err == nil {
		brands, recoErr := brandBuilder.BuildRecommend(ctx, request, []urlbuilderfactory.QueryOption{
			{
				Field: "brand",
				And:   true,
			},
		})
		if recoErr == nil && len(brands) > 0 {
			brandGroup := &entity.DynamicRecommendGroup{
				Group: "Gợi ý cùng nhãn hiệu",
				Data:  brands,
			}
			dataGroup = append(dataGroup, brandGroup)
		}
	}
	// year
	yearBuilder, err := urlbuilderfactory.BuilderFactory(entity.UrlKindYear, _self.shortlinkRepo)
	if err == nil {
		years, recoErr := yearBuilder.BuildRecommend(ctx, request, []urlbuilderfactory.QueryOption{
			{
				Field: "city",
				And:   false,
			},
			{
				Field: "category",
				And:   true,
			},
			{
				Field: "year",
				And:   true,
			},
		})
		if recoErr == nil && len(years) > 0 {
			yearGroup := &entity.DynamicRecommendGroup{
				Group: "Gợi ý cùng năm sản xuất",
				Data:  years,
			}
			dataGroup = append(dataGroup, yearGroup)
		}
	}
	return &entity.DynamicRecommend{
		Data: dataGroup,
	}, nil
}

func (_self *UrlService) CreateUrl(ctx context.Context, url entity.Url) error {
	var request domain.Url
	utils.Copy(&request, url)
	var newUrlId int64
	err := _self.db.RunWithTransaction(ctx,
		func(ctx context.Context, tx *gorm.DB) error {
			urlId, err := _self.urlRepo.CreateUrl(ctx, tx, request)
			if err != nil {
				return err
			}
			newUrlId = urlId
			return nil
		},
		func(ctx context.Context, tx *gorm.DB) error {
			var metadata []*domain.UrlMetadata
			for _, meta := range url.Metadata {
				var elem *domain.UrlMetadata
				utils.Copy(&elem, meta)
				elem.UrlId = newUrlId
				metadata = append(metadata, elem)
			}
			if err := _self.urlMetadataRepo.CreateUrlMetadata(ctx, tx, metadata); err != nil {
				return err
			}
			return nil
		})
	return err
}

func (_self *UrlService) UpdateUrl(ctx context.Context, url entity.Url) error {
	err := _self.db.RunWithTransaction(ctx,
		func(ctx context.Context, tx *gorm.DB) error {
			var request domain.Url
			utils.Copy(&request, url)
			if err := _self.urlRepo.UpdateUrl(ctx, tx, request); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, tx *gorm.DB) error {
			var metadata []*domain.UrlMetadata
			for _, meta := range url.Metadata {
				var elem *domain.UrlMetadata
				utils.Copy(&elem, meta)
				metadata = append(metadata, elem)
			}
			if err := _self.urlMetadataRepo.UpdateUrlMetadata(ctx, tx, metadata); err != nil {
				return err
			}
			return nil
		})
	return err
}

func (_self *UrlService) DeleteUrl(ctx context.Context, url entity.Url) error {
	err := _self.db.RunWithTransaction(ctx,
		func(ctx context.Context, tx *gorm.DB) error {
			var request domain.Url
			utils.Copy(&request, url)
			if err := _self.urlRepo.DeleteUrl(ctx, tx, request.Url); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, tx *gorm.DB) error {
			if err := _self.urlMetadataRepo.DeleteUrlMetadataById(ctx, tx, url.Id); err != nil {
				return err
			}
			return nil
		})
	return err
}

func (_self *UrlService) GetUrl(ctx context.Context, url string) (*entity.Url, error) {
	urlData, err := _self.urlRepo.GetUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	if urlData == nil {
		return nil, nil
	}
	metadata, err := _self.urlMetadataRepo.GetUrlMetadata(ctx, urlData.Id)
	if err != nil {
		return nil, err
	}
	resp := &entity.Url{}
	utils.Copy(&resp, urlData)
	if metadata != nil {
		var metadataEntity []*entity.UrlMetadata
		for _, meta := range metadata {
			var elem *entity.UrlMetadata
			utils.Copy(&elem, meta)
			metadataEntity = append(metadataEntity, elem)
		}
		resp.Metadata = metadataEntity
	}
	return resp, nil
}

func (_self *UrlService) GetUrls(ctx context.Context, offset, limit int) ([]*entity.Url, error) {
	urlDatas, err := _self.urlRepo.GetUrls(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, urlData := range urlDatas {
		ids = append(ids, urlData.Id)
	}
	metadata, err := _self.urlMetadataRepo.GetUrlMetadatas(ctx, ids)
	if err != nil {
		return nil, err
	}
	var resp []*entity.Url
	for _, urlData := range urlDatas {
		var metadataEntity []*entity.UrlMetadata
		for _, meta := range metadata {
			var elem *entity.UrlMetadata
			utils.Copy(&elem, meta)
			metadataEntity = append(metadataEntity, elem)
		}
		var urlElem *entity.Url
		utils.Copy(&resp, urlData)
		urlElem = &entity.Url{}
		utils.Copy(&urlElem, urlData)
		urlElem.Metadata = metadataEntity
		resp = append(resp, urlElem)
	}
	return resp, nil
}
