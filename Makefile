# Convenience wrappers over the docker compose commands documented in the
# README. Nothing here does anything you couldn't type by hand.

.PHONY: up down build logs ps seed promote-admin restart-backend restart-frontend clean

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

ps:
	docker compose ps

seed:
	docker compose run --rm backend ./seed

# make promote-admin EMAIL=you@example.com
promote-admin:
	@if [ -z "$(EMAIL)" ]; then \
		echo "usage: make promote-admin EMAIL=you@example.com"; \
		exit 1; \
	fi
	docker compose run --rm backend ./promote-admin $(EMAIL)

restart-backend:
	docker compose up -d --build backend

restart-frontend:
	docker compose up -d --build frontend

# Stops and removes containers AND the named postgres volume -- a full reset
# back to an empty database. Not run by anything else here.
clean:
	docker compose down -v
