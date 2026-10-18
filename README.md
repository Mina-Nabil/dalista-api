# dalista-api

Golang backend for Dalista — Gin + GORM + Postgres, following the Lumina
reference conventions. See `CLAUDE.md` for the full architecture and layering
rules, and `../../Claude outputs/` for the requirements/phases docs.

## Stack

Gin · GORM (`pgx` Postgres driver) · Goose SQL migrations · Sentry · `swaggo`
OpenAPI docs · `go-playground/validator` · `caarlos0/env` + `godotenv` config.

## Local development

```bash
cp .env.example .env            # fill in DB creds etc.
make tools                      # install goose, swag, air (one-time)
createdb dalista                # or point DB_* at an existing Postgres
make migrate-up                 # apply migrations
make run                        # or `make dev` for live reload
```

- API: `http://localhost:8080`
- Health: `GET /api/v1/health` (and ALB probe `GET /health`)
- Swagger UI (dev only): `http://localhost:8080/swagger/index.html`

## Common tasks (`make help`)

| Target | Purpose |
|--------|---------|
| `make run` / `make dev` | run the server (plain / live-reload) |
| `make build` | build the server binary |
| `make test` | run `go test ./test/...` (serial, real Postgres) |
| `make swagger` | regenerate `./docs` OpenAPI from handler annotations |
| `make migrate-up` / `migrate-down` / `migrate-status` | Goose migrations |

## Layout

```
main.go                     entrypoint (server.Initialize + server.Run)
config/                     env-driven Config, single GetConfig()
api/
  responses/                shared response envelopes
  server/
    initializer.go          Sentry/logger bootstrap
    runner.go               Gin engine, CORS, rate limit, routes
    logger/                 Sentry-backed structured logging
    middlewares/            rate limit (auth/permission land in Phase 1)
    routes/api.go           route groups per feature
    handlers/healthHandlers health probe
app/repository/
  database.go               GORM/Postgres singleton (GetDB/SetDB/Ping)
  migrations/*.sql          Goose migrations, timestamp-prefixed
test/                       tests mirror the prod package layout
docs/                       generated OpenAPI (swag init)
docker/staging/Dockerfile   multi-stage build (server + goose)
infra/                      Terraform (staging) — see infra/README.md
.github/workflows/          ci.yml (test+build), deploy.yml (ECS)
```

## CI/CD

- **CI** (`ci.yml`, on PR): spins up Postgres, runs Goose migrations + tests,
  verifies `docs/` is in sync, and does a build check.
- **CD** (`deploy.yml`, on push): `master` → staging, `prod` → production;
  builds + pushes the image to ECR, runs the migrate task, updates the ECS
  service (OIDC auth, no static AWS keys).
