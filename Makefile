.PHONY: help setup generate css css-watch build run dev test vet check import clean

help:
	@echo "setup      install build tools (templ CLI, Tailwind CLI)"
	@echo "generate   templ generate"
	@echo "css        build Tailwind CSS"
	@echo "build      generate + css + go build -> bin/farahdin"
	@echo "run        build and run (reads env, see .env.example)"
	@echo "dev        run with DEV_LOGIN=true on :8080"
	@echo "test       go test ./... (DB tests need TEST_POSTGRES_URL)"
	@echo "check      generate + vet + test"
	@echo "import     import a Convex export: make import EXPORT=convex-export.zip"

setup:
	go install github.com/a-h/templ/cmd/templ@v0.3.1020
	npm ci

generate:
	templ generate

css:
	npm run css

css-watch:
	npm run css:watch

build: generate css
	go build -trimpath -ldflags "-s -w" -o bin/farahdin ./cmd/app

run: build
	./bin/farahdin

dev: generate css
	APP_ENV=development DEV_LOGIN=true go run ./cmd/app

test:
	go test ./...

vet:
	go vet ./...

check: generate vet test

import:
	go run ./cmd/import-convex $(EXPORT)

clean:
	rm -rf bin web/static/css/app.css
