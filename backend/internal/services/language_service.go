package services

import (
	"context"
	"strings"

	"github.com/mrkaak/restaurant-api/internal/models"
	"github.com/mrkaak/restaurant-api/internal/repository"
)

// LanguageService manages locales and UI string bundles.
type LanguageService struct {
	langs *repository.LanguageRepo
}

func NewLanguageService(langs *repository.LanguageRepo) *LanguageService {
	return &LanguageService{langs: langs}
}

func (s *LanguageService) ListActive(ctx context.Context) ([]models.Language, error) {
	return s.langs.List(ctx, true)
}

func (s *LanguageService) ListAll(ctx context.Context) ([]models.Language, error) {
	return s.langs.List(ctx, false)
}

func (s *LanguageService) Bundle(ctx context.Context, locale string) (map[string]string, error) {
	return s.langs.Bundle(ctx, strings.ToLower(strings.TrimSpace(locale)))
}

func (s *LanguageService) Create(ctx context.Context, in LanguageInput) (*models.Language, error) {
	l := &models.Language{
		Code:       strings.ToLower(strings.TrimSpace(in.Code)),
		Name:       in.Name,
		NativeName: in.NativeName,
		IsRTL:      in.IsRTL,
		IsActive:   derefBool(in.IsActive, true),
		SortOrder:  in.SortOrder,
	}
	if err := s.langs.Create(ctx, l); err != nil {
		if repository.IsDuplicateKey(err) {
			return nil, ErrSlugTaken // reuse: code already exists
		}
		return nil, err
	}
	return l, nil
}

func (s *LanguageService) Update(ctx context.Context, id uint64, in LanguageInput) (*models.Language, error) {
	l, err := s.langs.FindByID(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	// Guard: don't deactivate the last active language or the default one.
	if in.IsActive != nil && !*in.IsActive {
		if l.IsDefault {
			return nil, ErrForbidden
		}
		if n, _ := s.langs.CountActive(ctx); n <= 1 {
			return nil, ErrForbidden
		}
	}
	l.Name = in.Name
	l.NativeName = in.NativeName
	l.IsRTL = in.IsRTL
	l.IsActive = derefBool(in.IsActive, l.IsActive)
	l.SortOrder = in.SortOrder
	if err := s.langs.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *LanguageService) SetDefault(ctx context.Context, id uint64) error {
	if err := s.langs.SetDefault(ctx, id); err != nil {
		return mapNotFound(err)
	}
	return nil
}

func (s *LanguageService) UpsertBundle(ctx context.Context, locale string, strs map[string]string) error {
	locale = strings.ToLower(strings.TrimSpace(locale))
	rows := make([]models.UIString, 0, len(strs))
	for k, v := range strs {
		rows = append(rows, models.UIString{Locale: locale, Key: k, Value: v})
	}
	return s.langs.UpsertStrings(ctx, rows)
}
