FROM golang:1.26.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.21 AS runtime
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app

FROM runtime AS api
COPY --from=builder /out/api /app/api
EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM runtime AS worker
COPY --from=builder /out/worker /app/worker
ENTRYPOINT ["/app/worker"]

FROM alpine:3.21 AS migrate
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
ENTRYPOINT ["goose", "-dir", "/migrations", "postgres"]
