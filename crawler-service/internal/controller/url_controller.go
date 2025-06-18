package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/service"

	// Import service

	crawlerv1 "github.com/namnv2496/crawler/pkg/generated/pkg/proto"
	"github.com/namnv2496/crawler/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UrlController struct {
	crawlerv1.UnimplementedUrlServiceServer
	urlService service.IUrlService
	ratelimter utils.IRateLimit
}

func NewUrlController(
	urlService service.IUrlService,
	ratelimter utils.IRateLimit,
) crawlerv1.UrlServiceServer {
	return &UrlController{
		urlService: urlService,
		ratelimter: ratelimter,
	}
}

func (_self *UrlController) CreateUrl(ctx context.Context, req *crawlerv1.CreateUrlRequest) (*crawlerv1.CreateUrlResponse, error) {
	// rate limit
	if err := _self.checkInserRateLimit(ctx, req.Url.Id); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
	}

	if req == nil || req.Url == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request or url is nil")
	}
	domainUrl := &entity.Url{
		Url:         req.Url.Url,
		Description: req.Url.Description,
		Queue:       req.Url.Queue,
		Domain:      req.Url.Domain,
		IsActive:    req.Url.IsActive,
		Method:      req.Url.Method,
	}

	id, err := _self.urlService.CreateUrl(ctx, domainUrl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create url: %v", err)
	}

	return &crawlerv1.CreateUrlResponse{
		Id:     strconv.FormatInt(id, 10),
		Status: "created",
	}, nil
}

func (_self *UrlController) GetUrls(ctx context.Context, req *crawlerv1.GetUrlsRequest) (*crawlerv1.GetUrlsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is nil")
	}
	// rate limit
	if err := _self.checkRateLimit(ctx, "test"); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	urls, err := _self.urlService.GetUrls(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get urls: %v", err)
	}

	protoUrls := make([]*crawlerv1.Url, len(urls))
	for i, url := range urls {
		protoUrls[i] = &crawlerv1.Url{
			Id:          fmt.Sprintf("%d", url.Id),
			Url:         url.Url,
			Method:      url.Method,
			Description: url.Description,
			Queue:       url.Queue,
			Domain:      url.Domain,
			IsActive:    url.IsActive,
			CreatedAt:   url.CreatedAt.String(),
			UpdatedAt:   url.UpdatedAt.String(),
		}
	}

	if len(protoUrls) == 0 {
		return &crawlerv1.GetUrlsResponse{
			Urls: []*crawlerv1.Url{},
		}, nil
	}
	return &crawlerv1.GetUrlsResponse{
		Urls: protoUrls,
	}, nil
}
func (_self *UrlController) UpdateUrl(ctx context.Context, req *crawlerv1.UpdateUrlRequest) (*crawlerv1.UpdateUrlResponse, error) {
	if req == nil || req.Url == nil || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request, url, or id is nil/empty")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	domainUrl := &entity.Url{
		Id:          id,
		Url:         req.Url.Url,
		Method:      req.Url.Method,
		Description: req.Url.Description,
		Queue:       req.Url.Queue,
		Domain:      req.Url.Domain,
		IsActive:    req.Url.IsActive,
	}

	err = _self.urlService.UpdateUrl(ctx, id, domainUrl)
	if err != nil {
		if errors.Is(err, errors.New("url not found")) { // Check for specific error from repository/service
			return nil, status.Errorf(codes.NotFound, "url with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to update url: %v", err)
	}

	return &crawlerv1.UpdateUrlResponse{
		Id:     req.Id,
		Status: "updated",
	}, nil
}

func (_self *UrlController) checkRateLimit(ctx context.Context, key string) error {
	// rate limit 10 request/ minute
	pass, err := _self.ratelimter.Allow(ctx, "query_url", key, utils.LimitPerMinute(10, 1))
	if err != nil {
		return err
	}
	if !pass {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (_self *UrlController) checkInserRateLimit(ctx context.Context, key string) error {
	// rate limit 50 request/ second
	pass, err := _self.ratelimter.Allow(ctx, "create_url", key, utils.LimitCustom(50, 1, time.Second))
	if err != nil {
		return err
	}
	if !pass {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}
