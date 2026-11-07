// promote-admin grants the admin role to an existing account, so the first
// admin can be created without editing the database by hand.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/ghostuf/reload/internal/config"
	"github.com/ghostuf/reload/internal/models"
	"github.com/ghostuf/reload/internal/repository"
	"gorm.io/gorm"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: promote-admin <email>")
		fmt.Fprintln(flag.CommandLine.Output(), "Grants the admin role to the account registered under <email>.")
	}
	flag.Parse()

	email := flag.Arg(0)
	if email == "" {
		flag.Usage()
		log.Fatal("missing email")
	}

	cfg := config.Load()

	db, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	users := repository.NewUserRepository(db)

	user, err := users.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("no account found for %s", email)
		}
		log.Fatalf("failed to look up %s: %v", email, err)
	}

	if user.Role == models.RoleAdmin {
		log.Printf("%s is already an admin", email)
		return
	}

	if err := users.UpdateRole(user.ID, models.RoleAdmin); err != nil {
		log.Fatalf("failed to promote %s: %v", email, err)
	}

	log.Printf("promoted %s to admin", email)
}
