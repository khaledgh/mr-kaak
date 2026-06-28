package services

import (
	"context"
	"strings"
	"time"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
)

// BannerService serves active banners (public) and admin CRUD.
type BannerService struct {
	banners *repository.BannerRepo
	baseURL string
}

func NewBannerService(banners *repository.BannerRepo, baseURL string) *BannerService {
	return &BannerService{
		banners: banners,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// ActiveNow returns banners currently within their schedule window.
func (s *BannerService) ActiveNow(ctx context.Context) ([]models.Banner, error) {
	list, err := s.banners.ActiveNow(ctx, time.Now())
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].ImageURL = s.resolveURL(list[i].ImageURL)
	}
	return list, nil
}

func (s *BannerService) List(ctx context.Context) ([]models.Banner, error) {
	list, err := s.banners.List(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].ImageURL = s.resolveURL(list[i].ImageURL)
	}
	return list, nil
}

func (s *BannerService) Create(ctx context.Context, in BannerInput) (*models.Banner, error) {
	b := s.toModel(in)
	if err := s.banners.Create(ctx, b); err != nil {
		return nil, err
	}
	b.ImageURL = s.resolveURL(b.ImageURL)
	return b, nil
}

func (s *BannerService) Update(ctx context.Context, id uint64, in BannerInput) (*models.Banner, error) {
	existing, err := s.banners.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	b := s.toModel(in)
	b.ID = existing.ID
	if err := s.banners.Update(ctx, b); err != nil {
		return nil, err
	}
	b.ImageURL = s.resolveURL(b.ImageURL)
	return b, nil
}

func (s *BannerService) Delete(ctx context.Context, id uint64) error {
	return s.banners.Delete(ctx, id)
}

func (s *BannerService) resolveURL(raw string) string {
	if raw == "" || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return s.baseURL + raw
}

func toRelativePath(url string) string {
	if idx := strings.Index(url, "/uploads/"); idx != -1 {
		return url[idx:]
	}
	return url
}

func (s *BannerService) toModel(in BannerInput) *models.Banner {
	return &models.Banner{
		Title:     strings.TrimSpace(in.Title),
		ImageURL:  toRelativePath(in.ImageURL),
		LinkURL:   in.LinkURL,
		SortOrder: in.SortOrder,
		StartsAt:  in.StartsAt,
		EndsAt:    in.EndsAt,
		IsActive:  derefBool(in.IsActive, true),
	}
}
