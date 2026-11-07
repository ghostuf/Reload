package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ghostuf/reload/internal/auth"
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

const (
	defaultReviewPageSize = 10
	maxReviewPageSize     = 50
	maxReviewBody         = 4000
)

var (
	// ErrReviewValidation covers a rejected review write: a rating outside
	// 1-5, or a missing or too-long body. The message carried on the error
	// says which.
	ErrReviewValidation = errors.New("invalid review")
	// ErrAlreadyReviewed is returned when a user tries to review a game they
	// have already reviewed -- the unique index enforces this, this just
	// gives it a name a handler can map onto 409.
	ErrAlreadyReviewed = errors.New("you have already reviewed this game")
	// ErrGameNotOwned is returned when a user tries to review a game that is
	// not in their library.
	ErrGameNotOwned = errors.New("you must own a game to review it")
	// ErrNotReviewAuthor is returned when someone other than a review's
	// author tries to edit it, or delete it without being an admin.
	ErrNotReviewAuthor = errors.New("you can only edit or delete your own review")
)

// ReviewRepo is the persistence slice ReviewService needs for reviews
// themselves.
type ReviewRepo interface {
	ListByGame(filter repository.ReviewFilter) ([]models.Review, int64, error)
	Create(review *models.Review) error
	GetByID(id uint) (*models.Review, error)
	Update(review *models.Review) error
	Delete(id uint) error
}

// LibraryOwnership is the one question CreateReview needs answered: does
// this user own this game.
type LibraryOwnership interface {
	HasGame(userID, gameID uint) (bool, error)
}

var (
	_ ReviewRepo       = (*repository.ReviewRepository)(nil)
	_ LibraryOwnership = (*repository.LibraryRepository)(nil)
	_ GameStore        = (*repository.GameRepository)(nil)
	_ auth.RoleStore   = (*repository.UserRepository)(nil)
)

type ReviewService struct {
	reviews ReviewRepo
	library LibraryOwnership
	games   GameStore
	roles   auth.RoleStore
}

func NewReviewService(reviews ReviewRepo, library LibraryOwnership, games GameStore, roles auth.RoleStore) *ReviewService {
	return &ReviewService{reviews: reviews, library: library, games: games, roles: roles}
}

type ListReviewsParams struct {
	GameID   uint
	Page     int
	PageSize int
}

type ListReviewsResult struct {
	Reviews    []models.Review
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

func (s *ReviewService) ListReviews(params ListReviewsParams) (ListReviewsResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > maxReviewPageSize {
		params.PageSize = defaultReviewPageSize
	}

	reviews, total, err := s.reviews.ListByGame(repository.ReviewFilter{
		GameID:   params.GameID,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return ListReviewsResult{}, err
	}

	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)
	if totalPages < 1 {
		totalPages = 1
	}

	return ListReviewsResult{
		Reviews:    reviews,
		Page:       params.Page,
		PageSize:   params.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

type CreateReviewParams struct {
	UserID uint
	GameID uint
	Rating int
	Body   string
}

// CreateReview enforces the one rule that makes a review mean something
// here: only someone who owns the game can leave one. The game lookup runs
// first so a bad game_id reads as "no such game" rather than the more
// alarming "you don't own it".
func (s *ReviewService) CreateReview(params CreateReviewParams) (*models.Review, error) {
	if err := validateReview(params.Rating, params.Body); err != nil {
		return nil, err
	}

	if _, err := s.games.GetByID(params.GameID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGameNotFound
		}
		return nil, fmt.Errorf("looking up game: %w", err)
	}

	owned, err := s.library.HasGame(params.UserID, params.GameID)
	if err != nil {
		return nil, fmt.Errorf("checking library ownership: %w", err)
	}
	if !owned {
		return nil, ErrGameNotOwned
	}

	review := &models.Review{
		UserID: params.UserID,
		GameID: params.GameID,
		Rating: params.Rating,
		Body:   strings.TrimSpace(params.Body),
	}

	if err := s.reviews.Create(review); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyReviewed
		}
		return nil, fmt.Errorf("creating review: %w", err)
	}

	return review, nil
}

type UpdateReviewParams struct {
	ReviewID uint
	UserID   uint
	Rating   int
	Body     string
}

// UpdateReview lets only the author change a review -- there is no admin
// override here, unlike delete: moderation removes a review, it does not
// rewrite someone else's words.
func (s *ReviewService) UpdateReview(params UpdateReviewParams) (*models.Review, error) {
	if err := validateReview(params.Rating, params.Body); err != nil {
		return nil, err
	}

	review, err := s.reviews.GetByID(params.ReviewID)
	if err != nil {
		return nil, err
	}
	if review.UserID != params.UserID {
		return nil, ErrNotReviewAuthor
	}

	review.Rating = params.Rating
	review.Body = strings.TrimSpace(params.Body)

	if err := s.reviews.Update(review); err != nil {
		return nil, err
	}

	return review, nil
}

// DeleteReview lets the author delete their own review, or an admin delete
// anyone's. It returns gorm.ErrRecordNotFound when the review does not
// exist.
func (s *ReviewService) DeleteReview(reviewID, userID uint) error {
	review, err := s.reviews.GetByID(reviewID)
	if err != nil {
		return err
	}

	if review.UserID != userID {
		role, err := s.roles.GetRole(userID)
		if err != nil {
			return fmt.Errorf("checking role: %w", err)
		}
		if role != models.RoleAdmin {
			return ErrNotReviewAuthor
		}
	}

	return s.reviews.Delete(reviewID)
}

func validateReview(rating int, body string) error {
	switch {
	case rating < 1 || rating > 5:
		return fmt.Errorf("%w: rating must be between 1 and 5", ErrReviewValidation)
	case strings.TrimSpace(body) == "":
		return fmt.Errorf("%w: body is required", ErrReviewValidation)
	case len([]rune(body)) > maxReviewBody:
		return fmt.Errorf("%w: body must be at most %d characters", ErrReviewValidation, maxReviewBody)
	default:
		return nil
	}
}
