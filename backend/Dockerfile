FROM golang:1.22-bookworm AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/queryforge-api ./cmd/api

FROM debian:bookworm-slim

WORKDIR /app
RUN useradd -u 10001 -m appuser && mkdir -p /app/storage && chown -R appuser:appuser /app
COPY --from=build /bin/queryforge-api /app/queryforge-api
COPY migrations /app/migrations
USER appuser

EXPOSE 8080
CMD ["/app/queryforge-api"]
