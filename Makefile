.PHONY: db seed api web test

# Create the local database if it does not exist yet (uses the local Postgres, user postgres)
db:
	@PGPASSWORD=postgres psql -h localhost -U postgres -tc "SELECT 1 FROM pg_database WHERE datname='astrophage'" | grep -q 1 || \
		PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE astrophage"

seed:
	cd backend && go run ./cmd/seed

api:
	cd backend && go run ./cmd/api

web:
	cd frontend && npm run dev

test:
	cd backend && go test ./...
	cd frontend && npm test
