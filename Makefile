.PHONY: run build migrate seed-sample

# Build all binaries into bin/.
build:
	go build -o bin/altern  ./app/main/altern
	go build -o bin/migrate ./app/main/migrate
	go build -o bin/seed    ./app/main/seed

# Run the app with hot reload (watches Go files and views/).
# Uses `air` via `go run` so no global install is required —
# the first run populates the module cache; subsequent runs are instant.
run:
	go run github.com/air-verse/air@latest -c .air.toml

# Apply database migrations (AutoMigrate all models).
migrate:
	go run ./app/main/migrate

# Seed the database with sample data (50 tools, 10 users, 10 lists, 20 categories).
# Bails out if tools already exist. Use FORCE=1 to wipe and re-seed.
seed-sample:
	go run ./app/main/seed $(if $(FORCE),--force,)
