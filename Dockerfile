# Stage 1: Build
FROM golang:alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/db/migration ./db/migration
COPY --from=builder /app/app.env .

RUN apk add --no-cache bash curl

EXPOSE 8080
CMD ["/app/main"]
