.PHONY: run migrate

# Run the app with hot reload (watches Go files and views/).
# Uses `air` via `go run` so no global install is required —
# the first run populates the module cache; subsequent runs are instant.
run:
	go run github.com/air-verse/air@latest -c .air.toml

# Apply database migrations (AutoMigrate all models).
migrate:
	go run ./app/main/migrate
