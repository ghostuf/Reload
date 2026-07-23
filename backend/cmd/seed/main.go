package main

import (
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

func cover(slug string) string {
	return "https://picsum.photos/seed/" + slug + "/600/800"
}

func gallery(slug string, n int) []models.GameImage {
	images := make([]models.GameImage, n)
	for i := 0; i < n; i++ {
		images[i] = models.GameImage{
			ImageURL:  "https://picsum.photos/seed/" + slug + "-gallery-" + string(rune('a'+i)) + "/1280/720",
			SortOrder: i,
		}
	}
	return images
}

type seedGame struct {
	Title         string
	Slug          string
	Description   string
	Developer     string
	Publisher     string
	ReleaseDate   time.Time
	Price         float64
	DiscountPrice *float64
	GenreNames    []string
	GalleryCount  int
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
			Price:         59.99,
			DiscountPrice: ptr(39.99),
			GenreNames:    []string{"Action", "RPG"},
			GalleryCount:  3,
		},
		{
			Title:        "Neon Drift Protocol",
			Slug:         "neon-drift-protocol",
			Description:  "High-speed arcade racing through a neon-drenched megacity where every shortcut is a risk worth taking.",
			Developer:    "Vector Sun",
			Publisher:    "Chrome Wave Interactive",
			ReleaseDate:  date("2024-07-19"),
			Price:        49.99,
			GenreNames:   []string{"Racing", "Action"},
			GalleryCount: 2,
		},
		{
			Title:         "Hollow Frequency",
			Slug:          "hollow-frequency",
			Description:   "A survival horror adventure aboard an abandoned radio station where the broadcasts never really stopped.",
			Developer:     "Static Bloom",
			Publisher:     "Quiet Room Games",
			ReleaseDate:   date("2022-10-31"),
			Price:         24.99,
			DiscountPrice: ptr(14.99),
			GenreNames:    []string{"Horror", "Adventure"},
			GalleryCount:  3,
		},
		{
			Title:        "Ironclad Dominion",
			Slug:         "ironclad-dominion",
			Description:  "Command sprawling empires across a fractured continent in this deep turn-based grand strategy game.",
			Developer:    "Bastion Logic",
			Publisher:    "Foundry Interactive",
			ReleaseDate:  date("2021-05-06"),
			Price:        39.99,
			GenreNames:   []string{"Strategy"},
			GalleryCount: 2,
		},
		{
			Title:         "Verdant Expanse",
			Slug:          "verdant-expanse",
			Description:   "A quiet, hand-painted exploration game about tending an overgrown world back to life.",
			Developer:     "Loam & Light",
			Publisher:     "Evergreen Collective",
			ReleaseDate:   date("2023-09-01"),
			Price:         19.99,
			DiscountPrice: ptr(9.99),
			GenreNames:    []string{"Adventure", "Indie"},
			GalleryCount:  3,
		},
		{
			Title:        "Fracture Point",
			Slug:         "fracture-point",
			Description:  "A fast, tactical multiplayer shooter built around destructible arenas and squad-based objectives.",
			Developer:    "Redline Syndicate",
			Publisher:    "Apex Circuit",
			ReleaseDate:  date("2024-02-20"),
			Price:        29.99,
			GenreNames:   []string{"Multiplayer", "Action"},
			GalleryCount: 2,
		},
		{
			Title:         "Cinderpath Chronicles",
			Slug:          "cinderpath-chronicles",
			Description:   "A story-driven indie RPG following a wandering lantern-keeper through a slowly dimming world.",
			Developer:     "Ember & Ash",
			Publisher:     "Lanternfly Games",
			ReleaseDate:   date("2020-11-12"),
			Price:         34.99,
			DiscountPrice: ptr(17.99),
			GenreNames:    []string{"RPG", "Indie"},
			GalleryCount:  3,
		},
		{
			Title:        "Deadlock Signal",
			Slug:         "deadlock-signal",
			Description:  "A cooperative horror shooter where four players must escort a signal through a city gone silent.",
			Developer:    "Null Sector",
			Publisher:    "Blackout Games",
			ReleaseDate:  date("2024-10-04"),
			Price:        27.99,
			GenreNames:   []string{"Horror", "Multiplayer"},
			GalleryCount: 2,
		},
		{
			Title:        "Skybound Wanderers",
			Slug:         "skybound-wanderers",
			Description:  "A cozy life-sim set among floating islands, where you build a home and befriend the sky itself.",
			Developer:    "Paper Kite Studio",
			Publisher:    "Windward Press",
			ReleaseDate:  date("2022-06-15"),
			Price:        22.99,
			GenreNames:   []string{"Adventure", "Simulation"},
			GalleryCount: 3,
		},
		{
			Title:         "Crimson Circuit GP",
			Slug:          "crimson-circuit-gp",
			Description:   "A full-throttle racing sim with a career-spanning championship and deep vehicle customization.",
			Developer:     "Torque Theory",
			Publisher:     "Chrome Wave Interactive",
			ReleaseDate:   date("2023-12-08"),
			Price:         44.99,
			DiscountPrice: ptr(29.99),
			GenreNames:    []string{"Racing", "Sports"},
			GalleryCount:  2,
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
			CoverImageURL: cover(sg.Slug),
			Genres:        genres,
			Images:        gallery(sg.Slug, sg.GalleryCount),
		}

		if err := db.Create(&game).Error; err != nil {
			log.Fatalf("failed to seed game %q: %v", sg.Title, err)
		}

		log.Printf("seeded %s (%d genres, %d images)", sg.Title, len(genres), sg.GalleryCount)
	}

	log.Println("seed complete: 10 games inserted")
}
