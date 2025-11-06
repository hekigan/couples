# Multi-stage build for Couple Card Game

# Stage 1: Build Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Stage 2: Build SASS
FROM node:18-alpine AS sass-builder

WORKDIR /app

# Copy SASS files
COPY sass/ ./sass/
COPY package.json ./

# Install SASS and compile
RUN npm install -g sass
RUN sass sass/main.scss static/css/main.css

# Stage 3: Final runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static files and templates
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=sass-builder /app/static/css ./static/css

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]

