#Build stage
FROM golang:1.25.3-alpine3.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

## Run Stage
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /app/main .
COPY --from=builder /app/db/migration ./db/migration

EXPOSE 8080

CMD ["./main"]

