# Local development image — hot reload via air, plus goose for migrations.
# Source is bind-mounted at runtime (see docker-compose.yml), so nothing is
# COPYed here; this image only carries the toolchain. The production build is
# docker/staging/Dockerfile (distroless, no shell/air).
FROM golang:1.26-alpine

RUN apk add --no-cache git

# Live-reload watcher and the migration runner.
RUN go install github.com/air-verse/air@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@v3.27.3

WORKDIR /app

EXPOSE 8080

# Overridden per-service in compose (api → air, migrate → goose up).
CMD ["air", "-c", ".air.docker.toml"]
