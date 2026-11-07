package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
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

// ErrGameValidation covers every rejected admin game write: a missing
// required field, or a negative price. The message carried on the error
// says which.
var ErrGameValidation = errors.New("invalid game")

// ErrSlugTaken is returned when a create or update would collide with
// another game's slug.
var ErrSlugTaken = errors.New("slug is already in use")

// ReviewStore is the read slice of review data GameService needs to attach a
// game's rating to its detail response.
type ReviewStore interface {
	Aggregate(gameID uint) (avgRating *float64, count int64, err error)
}

var _ ReviewStore = (*repository.ReviewRepository)(nil)

type ListGamesParams struct {
	Search   string
	Genre    string
	OnSale   bool
	Sort     string
	Page     int
	PageSize int
}

// GameListing is a game as it appears in the store grid: the game itself
// plus the one price a shopper actually pays, computed by effectivePrice so
// list, detail, and checkout never disagree about what a discount is worth.
type GameListing struct {
	models.Game
	EffectivePrice float64 `json:"effective_price"`
}

// GameDetail is a listing with the two things only a single game's page
// needs: its rating, aggregated rather than built from every review.
type GameDetail struct {
	models.Game
	EffectivePrice float64  `json:"effective_price"`
	AverageRating  *float64 `json:"average_rating"`
	ReviewCount    int64    `json:"review_count"`
}

type ListGamesResult struct {
	Games      []GameListing
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

type GameService struct {
	repo    *repository.GameRepository
	reviews ReviewStore
}

func NewGameService(repo *repository.GameRepository, reviews ReviewStore) *GameService {
	return &GameService{repo: repo, reviews: reviews}
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
		OnSale:   params.OnSale,
		Sort:     params.Sort,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return ListGamesResult{}, err
	}

	listings := make([]GameListing, len(games))
	for i, game := range games {
		listings[i] = GameListing{Game: game, EffectivePrice: effectivePrice(game)}
	}

	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)
	if totalPages < 1 {
		totalPages = 1
	}

	return ListGamesResult{
		Games:      listings,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *GameService) GetGame(id uint) (*GameDetail, error) {
	game, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	avg, count, err := s.reviews.Aggregate(id)
	if err != nil {
		return nil, fmt.Errorf("loading rating: %w", err)
	}

	return &GameDetail{
		Game:           *game,
		EffectivePrice: effectivePrice(*game),
		AverageRating:  avg,
		ReviewCount:    count,
	}, nil
}

// GameInput is the admin-supplied shape of a game: genres and gallery images
// as plain names and URLs, rather than the IDs and records an admin creating
// or editing a game would have no way to already know.
type GameInput struct {
	Title          string
	Slug           string
	Description    string
	Developer      string
	Publisher      string
	ReleaseDate    time.Time
	Price          float64
	DiscountPrice  *float64
	DiscountEndsAt *time.Time
	CoverImageURL  string
	GenreNames     []string
	ImageURLs      []string
}

// CreateGame validates the input, resolves genre names onto genre rows
// (creating any that don't exist yet), and inserts the game.
func (s *GameService) CreateGame(input GameInput) (*models.Game, error) {
	if err := validateGameInput(input); err != nil {
		return nil, err
	}

	genres, err := s.repo.ResolveGenres(input.GenreNames)
	if err != nil {
		return nil, err
	}

	game := &models.Game{
		Title:          strings.TrimSpace(input.Title),
		Slug:           strings.TrimSpace(input.Slug),
		Description:    input.Description,
		Developer:      input.Developer,
		Publisher:      input.Publisher,
		ReleaseDate:    input.ReleaseDate,
		Price:          input.Price,
		DiscountPrice:  input.DiscountPrice,
		DiscountEndsAt: input.DiscountEndsAt,
		CoverImageURL:  strings.TrimSpace(input.CoverImageURL),
		Genres:         genres,
		Images:         buildImages(input.ImageURLs),
	}

	if err := s.repo.Create(game); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("creating game: %w", err)
	}

	return game, nil
}

// UpdateGame validates the input and replaces a game's fields, genres, and
// images wholesale -- PUT semantics, not a partial patch. It returns
// gorm.ErrRecordNotFound when id does not exist.
func (s *GameService) UpdateGame(id uint, input GameInput) (*models.Game, error) {
	if err := validateGameInput(input); err != nil {
		return nil, err
	}

	genres, err := s.repo.ResolveGenres(input.GenreNames)
	if err != nil {
		return nil, err
	}

	game := &models.Game{
		ID:             id,
		Title:          strings.TrimSpace(input.Title),
		Slug:           strings.TrimSpace(input.Slug),
		Description:    input.Description,
		Developer:      input.Developer,
		Publisher:      input.Publisher,
		ReleaseDate:    input.ReleaseDate,
		Price:          input.Price,
		DiscountPrice:  input.DiscountPrice,
		DiscountEndsAt: input.DiscountEndsAt,
		CoverImageURL:  strings.TrimSpace(input.CoverImageURL),
		Genres:         genres,
		Images:         buildImages(input.ImageURLs),
	}

	if err := s.repo.Update(game); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	return s.repo.GetByID(id)
}

// DeleteGame soft-deletes a game. It returns gorm.ErrRecordNotFound when id
// does not exist.
func (s *GameService) DeleteGame(id uint) error {
	return s.repo.Delete(id)
}

func buildImages(urls []string) []models.GameImage {
	images := make([]models.GameImage, len(urls))
	for i, url := range urls {
		images[i] = models.GameImage{ImageURL: url, SortOrder: i}
	}
	return images
}

func validateGameInput(input GameInput) error {
	switch {
	case strings.TrimSpace(input.Title) == "":
		return fmt.Errorf("%w: title is required", ErrGameValidation)
	case strings.TrimSpace(input.Slug) == "":
		return fmt.Errorf("%w: slug is required", ErrGameValidation)
	case strings.TrimSpace(input.CoverImageURL) == "":
		return fmt.Errorf("%w: cover_image_url is required", ErrGameValidation)
	case input.Price < 0:
		return fmt.Errorf("%w: price cannot be negative", ErrGameValidation)
	case input.DiscountPrice != nil && *input.DiscountPrice < 0:
		return fmt.Errorf("%w: discount_price cannot be negative", ErrGameValidation)
	default:
		return nil
	}
}
