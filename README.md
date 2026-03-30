# Food Ordering App Backend

Go backend for a food ordering app using Gin, Postgres, Redis, and Docker Compose.

## Features

- Product APIs:
  - `GET /products`
  - `GET /products/:id`
- Order API:
  - `POST /orders`
- Health API:
  - `GET /health`
- Promo API:
  - `GET /checkpromo?code=FIFTYOFF`
- Coupon validation rules:
  - code length must be `8-10`
  - alphanumeric only
  - code must appear in at least `2` of the configured coupon base files
- Fixed discount for valid promo codes (default: `15%`)

## Stack

- Go (`gin-gonic/gin`)
- Postgres (`pgx/v5`)
- Redis (`go-redis/v9`)
- Docker Compose for local/dev VM setup

## Configuration

Environment variables (with defaults):

- `PORT` (default `8080`)
- `DATABASE_URL` (default local docker postgres binding)
- `REDIS_ADDR` (default `127.0.0.1:6379`)
- `CORS_ALLOWED_ORIGINS` (CSV list)
- `COUPON_BASE_URLS` (CSV list of `.gz` source URLs)
- `COUPON_DISCOUNT_PCT` (default `15`)
- `COUPON_CACHE_PATH` (default `./data/coupon_cache.bin`)

## Local Development (Makefile)

Common commands:

- `make help` - list commands
- `make deps` - download Go modules
- `make test` - run all tests
- `make run` - run API locally
- `make runapp` - start Postgres/Redis and run API locally
- `make build` - build binary to `./bin/server`
- `make migrate` - run migrations once
- `make up` - start docker compose stack
- `make up-build` - rebuild and start stack
- `make down` - stop stack
- `make ps` - show container status
- `make logs-api` - follow API logs
- `make health` - check health endpoint
- `make promo CODE=FIFTYOFF` - check promo endpoint

## Run with Docker Compose

Start:

```bash
docker-compose up -d --build
```

Check:

```bash
docker-compose ps
curl http://localhost:8080/health
curl "http://localhost:8080/checkpromo?code=FIFTYOFF"
```

Stop:

```bash
docker-compose down
```

## Phase 1 VM Deployment (single VM)

Recommended setup:

- one API container
- local Postgres + Redis
- restart policies enabled
- Postgres healthcheck + API depends on healthy DB
- persistent coupon cache volume mounted at `/var/lib/foodapp`

This repository's `docker-compose.yml` is already configured for this setup.

## Migrations and Seeding

- Migrations run at API startup.
- You can also run migration-only command:

```bash
go run ./cmd/migrate
```

- Product seed runs only when products table is empty.

## Notes on Coupon Loading

- On first run, app may take longer while downloading/parsing coupon sources.
- Progress indicators are logged in API logs.
- On subsequent runs, cache file (`COUPON_CACHE_PATH`) is used for faster startup.

