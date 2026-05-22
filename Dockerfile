FROM golang:1.26.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1
RUN go build -o /app/bin/org-structure-api ./cmd/org-structure-api

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/org-structure-api /app/bin/org-structure-api
COPY --from=builder /go/bin/goose /app/bin/goose
COPY migrations /app/migrations

COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

CMD ["/app/docker-entrypoint.sh"]