# ==============================================================================
# Stage 1: Build
# ==============================================================================
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the server + CLI binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/syncpermissions ./cmd/syncpermissions

# Install goose for running migrations as a one-off ECS task
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.3 && \
    cp "$(go env GOPATH)/bin/goose" /out/goose

# ==============================================================================
# Stage 2: Runtime
# ==============================================================================
FROM gcr.io/distroless/static-debian12

COPY --from=builder /out/server /server
COPY --from=builder /out/syncpermissions /syncpermissions
COPY --from=builder /out/goose /goose

# Migration files, run by the goose binary in a dedicated migrate task.
COPY --from=builder /build/app/repository/migrations/ /app/migrations/

EXPOSE 8080

CMD ["/server"]
