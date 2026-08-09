.PHONY: run build migrate seed-sample kill

# Build all binaries into bin/.
build:
	go build -o bin/altern  ./app/main/altern
	go build -o bin/migrate ./app/main/migrate
	go build -o bin/seed    ./app/main/seed

# Run the app with hot reload (watches Go files and views/).
# Uses `air` via `go run` so no global install is required —
# the first run populates the module cache; subsequent runs are instant.
# Always kills whatever is still bound to the app port first, so a
# terminal that got closed/killed without letting air clean up doesn't
# leave a ghost `tmp/altern` blocking the next run.
run: kill
	go run github.com/air-verse/air@latest -c .air.toml

# Free the app port and stop any leftover air/altern processes from a
# previous `make run` that didn't shut down cleanly.
kill:
	@PORT=$$(grep -E '^PORT=' .env 2>/dev/null | cut -d= -f2); \
	PORT=$${PORT:-1337}; \
	PIDS=$$(lsof -ti tcp:$$PORT 2>/dev/null); \
	if [ -n "$$PIDS" ]; then \
		echo "Killing process(es) on port $$PORT: $$PIDS"; \
		kill -9 $$PIDS 2>/dev/null || true; \
	fi; \
	pkill -f "$(CURDIR)/tmp/altern" 2>/dev/null || true; \
	true

# Apply database migrations (AutoMigrate all models).
migrate:
	go run ./app/main/migrate

# Seed the database with sample data (50 tools, 10 users, 10 lists, 20 categories).
# Bails out if tools already exist. Use FORCE=1 to wipe and re-seed.
seed-sample:
	go run ./app/main/seed $(if $(FORCE),--force,)
