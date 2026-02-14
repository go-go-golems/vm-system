.PHONY: dev-backend dev-frontend frontend-install frontend-check frontend-build web-generate build

dev-backend:
	GOWORK=off go run ./cmd/vm-system serve --listen 127.0.0.1:3210

dev-frontend:
	pnpm -C ui dev

frontend-install:
	pnpm -C ui install --frozen-lockfile

frontend-check:
	pnpm -C ui check

frontend-build:
	pnpm -C ui run build

web-generate:
	go generate ./internal/web

build:
	go generate ./internal/web
	GOWORK=off go build -tags embed -o vm-system ./cmd/vm-system
