FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o /karthub ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates sqlite

WORKDIR /app
COPY --from=builder /karthub .

RUN mkdir -p /app/data /app/data/uploads

EXPOSE 8080

ENV KARTHUB_DB_PATH=/app/data/karthub.db

VOLUME ["/app/data"]

ENTRYPOINT ["/app/karthub"]
