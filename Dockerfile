# Build stage
FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

COPY --from=builder /build/app /app/app
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /build/scripts/migrate.sh /app/scripts/migrate.sh

COPY --from=builder /build/internal/config/config.yaml /app/internal/config/config.yaml

COPY --from=builder /build/migrations /app/migrations

COPY --from=builder /build/scripts/migrate.sh /app/scripts/migrate.sh
RUN chmod +x /app/scripts/migrate.sh


CMD ["/app/app"]
