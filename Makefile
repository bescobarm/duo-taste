.PHONY: api web build test

# Run the Go API on :8080
api:
	cd backend && go run ./cmd/api

# Run the Vue dev server on :5173 (proxies /api to the Go API)
web:
	cd frontend && npm run dev

build:
	cd backend && go build ./...
	cd frontend && npm run build

test:
	cd backend && go test ./...
