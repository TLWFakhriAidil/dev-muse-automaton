# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Copy package files and npm configuration
COPY package*.json .npmrc ./

# Install dependencies (using npm install instead of npm ci since package-lock.json might not be available)
RUN npm install --production=false

# Copy source code and configuration files needed for build
COPY src/ ./src/
COPY public/ ./public/
COPY index.html ./
COPY vite.config.ts ./
COPY tsconfig*.json ./
COPY tailwind.config.ts ./
COPY postcss.config.js ./
COPY components.json ./
COPY eslint.config.js ./

# Copy build timestamp for cache invalidation
COPY BUILD_TIMESTAMP ./

# Clean any existing build and node cache AGGRESSIVELY
RUN rm -rf dist/ node_modules/.vite node_modules/.cache .vite/

# Set build arguments for Vite environment variables
# These will be passed from Railway environment variables during build
# Default to production values if not provided
ARG VITE_SUPABASE_URL=https://bjnjucwpwdzgsnqmpmff.supabase.co
ARG VITE_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImJqbmp1Y3dwd2R6Z3NucW1wbWZmIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjA0OTk1MzksImV4cCI6MjA3NjA3NTUzOX0.vw1rOUqYWFkPNDwTdEgIfsCO9pyvTsFKaXHq3RcRTNU

# Export as environment variables for Vite build
ENV VITE_SUPABASE_URL=$VITE_SUPABASE_URL
ENV VITE_SUPABASE_ANON_KEY=$VITE_SUPABASE_ANON_KEY

# Debug: Print environment variables to verify they're set
RUN echo "Building frontend with VITE_SUPABASE_URL=$VITE_SUPABASE_URL"

# Build the React application with environment variables
RUN npm run build

# Backend build stage - Using Go 1.23 which is the latest stable version available in Docker Hub
FROM golang:1.23-alpine AS backend-builder

# Install minimal dependencies for CGO-free builds
RUN apk add --no-cache git ca-certificates tzdata

# Set CGO environment variables for static builds
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /app/bin/server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies including bash
RUN apk add --no-cache ca-certificates tzdata wget bash

# Create app directory
RUN mkdir -p /app

# Copy binary from backend builder
COPY --from=backend-builder /app/bin/server /app/server

# Copy built React application from frontend builder
COPY --from=frontend-builder /app/dist /app/dist

# Copy templates and static files (fallback)
COPY --from=backend-builder /src/templates /app/templates
COPY --from=backend-builder /src/static /app/static

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8080

# Set default environment variables (Railway will override PORT at runtime)
ENV PORT=8080
ENV APP_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

# Run the application directly
CMD ["/app/server"]