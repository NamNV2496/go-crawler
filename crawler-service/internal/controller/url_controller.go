package controller

import (
	"context"
	"errors"
	"strconv"

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/service"

	// Import service
	crawlerv1 "github.com/namnv2496/crawler/pkg/generated/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UrlController struct {
	crawlerv1.UnimplementedUrlServiceServer
	urlService service.IUrlService
}

func NewUrlController(
	urlService service.IUrlService,
) crawlerv1.UrlServiceServer {
	return &UrlController{
		urlService: urlService,
	}
}

func (u *UrlController) CreateUrl(ctx context.Context, req *crawlerv1.CreateUrlRequest) (*crawlerv1.CreateUrlResponse, error) {
	if req == nil || req.Url == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request or url is nil")
	}

	// Map protobuf Url to domain Url
	domainUrl := &domain.Url{
		Url:         req.Url.Url,
		Description: req.Url.Description,
		Queue:       req.Url.Queue,
		Domain:      req.Url.Domain,
		IsActive:    req.Url.IsActive,
		Method:      req.Url.Method,
		// ID, CreatedAt, UpdatedAt will be set by the service/repository
	}

	id, err := u.urlService.CreateUrl(ctx, domainUrl)
	if err != nil {
		// Handle specific errors if needed, otherwise return a generic internal error
		return nil, status.Errorf(codes.Internal, "failed to create url: %v", err)
	}

	return &crawlerv1.CreateUrlResponse{
		Id:     strconv.FormatInt(id, 10),
		Status: "created", // Or a more descriptive status
	}, nil
}

func (u *UrlController) GetUrls(ctx context.Context, req *crawlerv1.GetUrlsRequest) (*crawlerv1.GetUrlsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is nil")
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	urls, err := u.urlService.GetUrls(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get urls: %v", err)
	}

	// Map domain Urls to protobuf Urls
	protoUrls := make([]*crawlerv1.Url, len(urls))
	for i, url := range urls {
		protoUrls[i] = &crawlerv1.Url{
			Id:          strconv.FormatInt(url.Id, 10), // Convert int64 to string for the ID field in the response struc
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

	return &crawlerv1.GetUrlsResponse{
		Urls: protoUrls,
	}, nil
}

func (u *UrlController) UpdateUrl(ctx context.Context, req *crawlerv1.UpdateUrlRequest) (*crawlerv1.UpdateUrlResponse, error) {
	if req == nil || req.Url == nil || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request, url, or id is nil/empty")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	domainUrl := &domain.Url{
		Id:          id, // Convert string ID to int64
		Url:         req.Url.Url,
		Method:      req.Url.Method,
		Description: req.Url.Description,
		Queue:       req.Url.Queue,
		Domain:      req.Url.Domain,
		IsActive:    req.Url.IsActive,
	}

	err = u.urlService.UpdateUrl(ctx, id, domainUrl)
	if err != nil {
		if errors.Is(err, errors.New("url not found")) { // Check for specific error from repository/service
			return nil, status.Errorf(codes.NotFound, "url with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to update url: %v", err)
	}

	return &crawlerv1.UpdateUrlResponse{
		Id:     req.Id,
		Status: "updated", // Or a more descriptive status
	}, nil
}
