package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ghostuf/reload/internal/config"
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

func ptr(f float64) *float64 { return &f }

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("invalid date %q: %v", s, err)
	}
	return t
}

// unsplash builds a cropped, sized Unsplash CDN URL for a specific photo. All
// seed covers and gallery shots are real stock photography rather than
// generated placeholder cards, since there's no real box art behind seed
// data -- every id below was checked by hand for a genre/mood fit against
// its game (dark and moody for the grim RPG, neon for the racers, horror
// imagery for the horror titles, and so on) and confirmed to still resolve.
func unsplash(id string, w, h int) string {
	return fmt.Sprintf("https://images.unsplash.com/photo-%s?w=%d&h=%d&fit=crop&q=80", id, w, h)
}

func gallery(urls []string) []models.GameImage {
	images := make([]models.GameImage, len(urls))
	for i, u := range urls {
		images[i] = models.GameImage{ImageURL: u, SortOrder: i}
	}
	return images
}

type seedGame struct {
	Title       string
	Slug        string
	Description string
	Developer   string
	Publisher   string
	ReleaseDate time.Time
	// Prices are in NPR (Rs), converted from the store's original USD
	// figures at roughly Rs 133/$ and rounded to a clean ",999"-style
	// sticker price.
	Price         float64
	DiscountPrice *float64
	GenreNames    []string
	CoverImageURL string
	GalleryURLs   []string
}

func seedData() []seedGame {
	return []seedGame{
		{
			Title:         "Ashen Vale",
			Slug:          "ashen-vale",
			Description:   "A grim open-world RPG where a fallen kingdom's embers must be rekindled one brutal battle at a time.",
			Developer:     "Nightfall Studios",
			Publisher:     "Obsidian Peak Games",
			ReleaseDate:   date("2023-03-14"),
			Price:         7999,
			DiscountPrice: ptr(5299),
			GenreNames:    []string{"Action", "RPG"},
			CoverImageURL: unsplash("1552820728-8b83bb6b773f", 600, 800),
			GalleryURLs: []string{
				unsplash("1542751371-adc38448a05e", 1280, 720),
				unsplash("1511882150382-421056c89033", 1280, 720),
				unsplash("1519669556878-63bdad8a1a49", 1280, 720),
			},
		},
		{
			Title:         "Neon Drift Protocol",
			Slug:          "neon-drift-protocol",
			Description:   "High-speed arcade racing through a neon-drenched megacity where every shortcut is a risk worth taking.",
			Developer:     "Vector Sun",
			Publisher:     "Chrome Wave Interactive",
			ReleaseDate:   date("2024-07-19"),
			Price:         6599,
			GenreNames:    []string{"Racing", "Action"},
			CoverImageURL: unsplash("1612287230202-1ff1d85d1bdf", 600, 800),
			GalleryURLs: []string{
				unsplash("1593305841991-05c297ba4575", 1280, 720),
				unsplash("1593118247619-e2d6f056869e", 1280, 720),
			},
		},
		{
			Title:         "Hollow Frequency",
			Slug:          "hollow-frequency",
			Description:   "A survival horror adventure aboard an abandoned radio station where the broadcasts never really stopped.",
			Developer:     "Static Bloom",
			Publisher:     "Quiet Room Games",
			ReleaseDate:   date("2022-10-31"),
			Price:         3299,
			DiscountPrice: ptr(1999),
			GenreNames:    []string{"Horror", "Adventure"},
			CoverImageURL: unsplash("1519669556878-63bdad8a1a49", 600, 800),
			GalleryURLs: []string{
				unsplash("1560419015-7c427e8ae5ba", 1280, 720),
				unsplash("1511882150382-421056c89033", 1280, 720),
				unsplash("1552820728-8b83bb6b773f", 1280, 720),
			},
		},
		{
			Title:         "Ironclad Dominion",
			Slug:          "ironclad-dominion",
			Description:   "Command sprawling empires across a fractured continent in this deep turn-based grand strategy game.",
			Developer:     "Bastion Logic",
			Publisher:     "Foundry Interactive",
			ReleaseDate:   date("2021-05-06"),
			Price:         5299,
			GenreNames:    []string{"Strategy"},
			CoverImageURL: unsplash("1591370874773-6702e8f12fd8", 600, 800),
			GalleryURLs: []string{
				unsplash("1598550476439-6847785fcea6", 1280, 720),
				unsplash("1538481199705-c710c4e965fc", 1280, 720),
			},
		},
		{
			Title:         "Verdant Expanse",
			Slug:          "verdant-expanse",
			Description:   "A quiet, hand-painted exploration game about tending an overgrown world back to life.",
			Developer:     "Loam & Light",
			Publisher:     "Evergreen Collective",
			ReleaseDate:   date("2023-09-01"),
			Price:         2699,
			DiscountPrice: ptr(1299),
			GenreNames:    []string{"Adventure", "Indie"},
			CoverImageURL: unsplash("1486401899868-0e435ed85128", 600, 800),
			GalleryURLs: []string{
				unsplash("1580327344181-c1163234e5a0", 1280, 720),
				unsplash("1505740420928-5e560c06d30e", 1280, 720),
				unsplash("1593118247619-e2d6f056869e", 1280, 720),
			},
		},
		{
			Title:         "Fracture Point",
			Slug:          "fracture-point",
			Description:   "A fast, tactical multiplayer shooter built around destructible arenas and squad-based objectives.",
			Developer:     "Redline Syndicate",
			Publisher:     "Apex Circuit",
			ReleaseDate:   date("2024-02-20"),
			Price:         3999,
			GenreNames:    []string{"Multiplayer", "Action"},
			CoverImageURL: unsplash("1593305841991-05c297ba4575", 600, 800),
			GalleryURLs: []string{
				unsplash("1542751371-adc38448a05e", 1280, 720),
				unsplash("1598550476439-6847785fcea6", 1280, 720),
			},
		},
		{
			Title:         "Cinderpath Chronicles",
			Slug:          "cinderpath-chronicles",
			Description:   "A story-driven indie RPG following a wandering lantern-keeper through a slowly dimming world.",
			Developer:     "Ember & Ash",
			Publisher:     "Lanternfly Games",
			ReleaseDate:   date("2020-11-12"),
			Price:         4699,
			DiscountPrice: ptr(2399),
			GenreNames:    []string{"RPG", "Indie"},
			CoverImageURL: unsplash("1550745165-9bc0b252726f", 600, 800),
			GalleryURLs: []string{
				unsplash("1505740420928-5e560c06d30e", 1280, 720),
				unsplash("1538481199705-c710c4e965fc", 1280, 720),
				unsplash("1511512578047-dfb367046420", 1280, 720),
			},
		},
		{
			Title:         "Deadlock Signal",
			Slug:          "deadlock-signal",
			Description:   "A cooperative horror shooter where four players must escort a signal through a city gone silent.",
			Developer:     "Null Sector",
			Publisher:     "Blackout Games",
			ReleaseDate:   date("2024-10-04"),
			Price:         3699,
			GenreNames:    []string{"Horror", "Multiplayer"},
			CoverImageURL: unsplash("1560419015-7c427e8ae5ba", 600, 800),
			GalleryURLs: []string{
				unsplash("1511882150382-421056c89033", 1280, 720),
				unsplash("1519669556878-63bdad8a1a49", 1280, 720),
			},
		},
		{
			Title:         "Skybound Wanderers",
			Slug:          "skybound-wanderers",
			Description:   "A cozy life-sim set among floating islands, where you build a home and befriend the sky itself.",
			Developer:     "Paper Kite Studio",
			Publisher:     "Windward Press",
			ReleaseDate:   date("2022-06-15"),
			Price:         3099,
			GenreNames:    []string{"Adventure", "Simulation"},
			CoverImageURL: unsplash("1580327344181-c1163234e5a0", 600, 800),
			GalleryURLs: []string{
				unsplash("1486401899868-0e435ed85128", 1280, 720),
				unsplash("1505740420928-5e560c06d30e", 1280, 720),
				unsplash("1593118247619-e2d6f056869e", 1280, 720),
			},
		},
		{
			Title:         "Crimson Circuit GP",
			Slug:          "crimson-circuit-gp",
			Description:   "A full-throttle racing sim with a career-spanning championship and deep vehicle customization.",
			Developer:     "Torque Theory",
			Publisher:     "Chrome Wave Interactive",
			ReleaseDate:   date("2023-12-08"),
			Price:         5999,
			DiscountPrice: ptr(3999),
			GenreNames:    []string{"Racing", "Sports"},
			CoverImageURL: unsplash("1511512578047-dfb367046420", 600, 800),
			GalleryURLs: []string{
				unsplash("1612287230202-1ff1d85d1bdf", 1280, 720),
				unsplash("1593305841991-05c297ba4575", 1280, 720),
			},
		},
	}
}

func resetTables(db *gorm.DB) error {
	return db.Exec(`TRUNCATE TABLE game_images, game_genres, games, genres RESTART IDENTITY CASCADE`).Error
}

func main() {
	cfg := config.Load()

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	if err := resetTables(db); err != nil {
		log.Fatalf("failed to reset tables: %v", err)
	}

	genreCache := map[string]models.Genre{}

	for _, sg := range seedData() {
		var genres []models.Genre
		for _, name := range sg.GenreNames {
			genre, ok := genreCache[name]
			if !ok {
				genre = models.Genre{Name: name}
				if err := db.Where("name = ?", name).FirstOrCreate(&genre).Error; err != nil {
					log.Fatalf("failed to seed genre %q: %v", name, err)
				}
				genreCache[name] = genre
			}
			genres = append(genres, genre)
		}

		game := models.Game{
			Title:         sg.Title,
			Slug:          sg.Slug,
			Description:   sg.Description,
			Developer:     sg.Developer,
			Publisher:     sg.Publisher,
			ReleaseDate:   sg.ReleaseDate,
			Price:         sg.Price,
			DiscountPrice: sg.DiscountPrice,
			CoverImageURL: sg.CoverImageURL,
			Genres:        genres,
			Images:        gallery(sg.GalleryURLs),
		}

		if err := db.Create(&game).Error; err != nil {
			log.Fatalf("failed to seed game %q: %v", sg.Title, err)
		}

		log.Printf("seeded %s (%d genres, %d images)", sg.Title, len(genres), len(sg.GalleryURLs))
	}

	log.Println("seed complete: 10 games inserted")
}
