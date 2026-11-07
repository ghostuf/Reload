package services

import (
	"errors"
	"testing"

	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

// fakeReviewRepo is an in-memory ReviewRepo that reproduces the one
// behaviour CreateReview depends on: a duplicate (user_id, game_id) surfaces
// as gorm.ErrDuplicatedKey.
type fakeReviewRepo struct {
	reviews map[uint]*models.Review
	nextID  uint
}

func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{reviews: make(map[uint]*models.Review), nextID: 1}
}

func (r *fakeReviewRepo) ListByGame(filter repository.ReviewFilter) ([]models.Review, int64, error) {
	var out []models.Review
	for _, rev := range r.reviews {
		if rev.GameID == filter.GameID {
			out = append(out, *rev)
		}
	}
	return out, int64(len(out)), nil
}

func (r *fakeReviewRepo) Create(review *models.Review) error {
	for _, existing := range r.reviews {
		if existing.UserID == review.UserID && existing.GameID == review.GameID {
			return gorm.ErrDuplicatedKey
		}
	}
	review.ID = r.nextID
	r.nextID++
	stored := *review
	r.reviews[review.ID] = &stored
	return nil
}

func (r *fakeReviewRepo) GetByID(id uint) (*models.Review, error) {
	rev, ok := r.reviews[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	copied := *rev
	return &copied, nil
}

func (r *fakeReviewRepo) Update(review *models.Review) error {
	if _, ok := r.reviews[review.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	stored := *review
	r.reviews[review.ID] = &stored
	return nil
}

func (r *fakeReviewRepo) Delete(id uint) error {
	if _, ok := r.reviews[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.reviews, id)
	return nil
}

// fakeLibrary answers ownership questions from a fixed userID -> gameID set,
// standing in for whichever games a test wants the user to already own.
type fakeLibrary struct {
	owned map[uint]map[uint]bool
}

func (l *fakeLibrary) HasGame(userID, gameID uint) (bool, error) {
	return l.owned[userID][gameID], nil
}

// fakeGames reports a game as existing unless its ID is listed as missing,
// standing in for the catalog lookup CreateReview does before checking
// ownership.
type fakeGames struct {
	missing map[uint]bool
}

func (g *fakeGames) GetByID(id uint) (*models.Game, error) {
	if g.missing[id] {
		return nil, gorm.ErrRecordNotFound
	}
	return &models.Game{ID: id}, nil
}

// fakeRoles answers role lookups from a fixed set of admin user IDs.
type fakeRoles struct {
	admins map[uint]bool
}

func (r *fakeRoles) GetRole(userID uint) (models.UserRole, error) {
	if r.admins[userID] {
		return models.RoleAdmin, nil
	}
	return models.RoleUser, nil
}

// TestCreateReviewEligibility is the rule that makes a review mean
// something here: only someone who owns the game can leave one.
func TestCreateReviewEligibility(t *testing.T) {
	t.Run("owns the game", func(t *testing.T) {
		reviews := newFakeReviewRepo()
		s := NewReviewService(
			reviews,
			&fakeLibrary{owned: map[uint]map[uint]bool{7: {1: true}}},
			&fakeGames{missing: map[uint]bool{}},
			&fakeRoles{},
		)

		review, err := s.CreateReview(CreateReviewParams{UserID: 7, GameID: 1, Rating: 5, Body: "Great game"})
		if err != nil {
			t.Fatalf("CreateReview: %v", err)
		}
		if review.ID == 0 {
			t.Error("review was not assigned an ID")
		}
		if len(reviews.reviews) != 1 {
			t.Errorf("stored reviews = %d, want 1", len(reviews.reviews))
		}
	})

	t.Run("does not own the game", func(t *testing.T) {
		reviews := newFakeReviewRepo()
		s := NewReviewService(
			reviews,
			&fakeLibrary{owned: map[uint]map[uint]bool{}},
			&fakeGames{missing: map[uint]bool{}},
			&fakeRoles{},
		)

		_, err := s.CreateReview(CreateReviewParams{UserID: 7, GameID: 1, Rating: 5, Body: "Great game"})
		if !errors.Is(err, ErrGameNotOwned) {
			t.Errorf("error = %v, want ErrGameNotOwned", err)
		}
		if len(reviews.reviews) != 0 {
			t.Error("a review was created despite the game not being owned")
		}
	})

	t.Run("game does not exist", func(t *testing.T) {
		// A bad game_id must read as "no such game", not the more alarming
		// "you don't own it" -- the game lookup runs before the ownership
		// check for exactly this reason.
		s := NewReviewService(
			newFakeReviewRepo(),
			&fakeLibrary{owned: map[uint]map[uint]bool{7: {1: true}}},
			&fakeGames{missing: map[uint]bool{1: true}},
			&fakeRoles{},
		)

		_, err := s.CreateReview(CreateReviewParams{UserID: 7, GameID: 1, Rating: 5, Body: "Great game"})
		if !errors.Is(err, ErrGameNotFound) {
			t.Errorf("error = %v, want ErrGameNotFound", err)
		}
	})
}
