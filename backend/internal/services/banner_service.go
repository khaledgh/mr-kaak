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
}

func NewBannerService(banners *repository.BannerRepo) *BannerService {
	return &BannerService{banners: banners}
}

// ActiveNow returns banners currently within their schedule window.
func (s *BannerService) ActiveNow(ctx context.Context) ([]models.Banner, error) {
	return s.banners.ActiveNow(ctx, time.Now())
}

func (s *BannerService) List(ctx context.Context) ([]models.Banner, error) {
	return s.banners.List(ctx)
}

func (s *BannerService) Create(ctx context.Context, in BannerInput) (*models.Banner, error) {
	b := in.toModel()
	if err := s.banners.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BannerService) Update(ctx context.Context, id uint64, in BannerInput) (*models.Banner, error) {
	existing, err := s.banners.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	b := in.toModel()
	b.ID = existing.ID
	if err := s.banners.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *BannerService) Delete(ctx context.Context, id uint64) error {
	return s.banners.Delete(ctx, id)
}

func (in BannerInput) toModel() *models.Banner {
	return &models.Banner{
		Title:     strings.TrimSpace(in.Title),
		ImageURL:  in.ImageURL,
		LinkURL:   in.LinkURL,
		SortOrder: in.SortOrder,
		StartsAt:  in.StartsAt,
		EndsAt:    in.EndsAt,
		IsActive:  derefBool(in.IsActive, true),
	}
}
