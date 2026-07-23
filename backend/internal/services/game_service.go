package services

import (
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
)

const (
	defaultPageSize = 12
	maxPageSize     = 50
)

var validSorts = map[string]bool{
	"newest":       true,
	"price_asc":    true,
	"price_desc":   true,
	"title_asc":    true,
	"title_desc":   true,
	"release_date": true,
}

type ListGamesParams struct {
	Search   string
	Genre    string
	Sort     string
	Page     int
	PageSize int
}

type ListGamesResult struct {
	Games      []models.Game
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type GameService struct {
	repo *repository.GameRepository
}

func NewGameService(repo *repository.GameRepository) *GameService {
	return &GameService{repo: repo}
}

func (s *GameService) ListGames(params ListGamesParams) (ListGamesResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > maxPageSize {
		params.PageSize = defaultPageSize
	}
	if !validSorts[params.Sort] {
		params.Sort = "newest"
	}

	games, total, err := s.repo.List(repository.GameFilter{
		Search:   params.Search,
		Genre:    params.Genre,
		Sort:     params.Sort,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return ListGamesResult{}, err
	}

	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)
	if totalPages < 1 {
		totalPages = 1
	}

	return ListGamesResult{
		Games:      games,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *GameService) GetGame(id uint) (*models.Game, error) {
	return s.repo.GetByID(id)
}
