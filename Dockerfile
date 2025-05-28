# Base Golang image
FROM golang:1.24.3-alpine AS base

# Development stage - doesn't copy source code, expects bind mount
FROM base AS development
WORKDIR /app
# Install air for development only
RUN go install github.com/air-verse/air@latest
EXPOSE 3000
# Create .air.toml if it doesn't exist
RUN if [ ! -f .air.toml ]; then air init; fi
CMD ["air"]

# Production build stage
FROM base AS build-stage
WORKDIR /app
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build-stage /app/main .
EXPOSE 3000
CMD ["./main"]