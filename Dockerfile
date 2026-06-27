FROM alpine:3.18
WORKDIR /app

# Copy the pre-built binary from GitHub Actions
COPY server-app ./main

# Copy other necessary files
COPY db/migration ./db/migration
COPY app.env .

RUN apk add --no-cache bash curl

EXPOSE 8080
CMD ["/app/main"]
