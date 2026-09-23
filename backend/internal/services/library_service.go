package services

import (
	"fmt"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
)

type LibraryStore interface {
	ListByUser(userID uint) ([]models.LibraryItem, error)
}

var _ LibraryStore = (*repository.LibraryRepository)(nil)

type LibraryService struct {
	library LibraryStore
}

func NewLibraryService(library LibraryStore) *LibraryService {
	return &LibraryService{library: library}
}

func (s *LibraryService) GetLibrary(userID uint) ([]models.LibraryItem, error) {
	items, err := s.library.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("loading library: %w", err)
	}
	return items, nil
}
