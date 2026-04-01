.PHONY: build up down logs test migrate clean

# Build the Nakama Go plugin
build:
	cd nakama && docker build -t realmweaver-nakama .

# Start all services
up: build
	docker compose up -d

# Stop all services
down:
	docker compose down

# View logs (follow)
logs:
	docker compose logs -f nakama

# Run Go tests
test:
	cd nakama && go test ./...

# Run migrations against local PostgreSQL
migrate:
	PGPASSWORD=realmweaver_dev psql -h localhost -p 5433 -U realmweaver -d realmweaver_world -f migrations/001_world_tables.sql

# Remove all data volumes
clean:
	docker compose down -v
