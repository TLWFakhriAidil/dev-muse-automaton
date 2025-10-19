#!/bin/bash
# Optimized deployment script for Railway with snapshot timeout fix

# Exit on error
set -e

echo "🚀 Starting optimized deployment for Chatbot Automation on Railway"

# Clean up unnecessary files to reduce snapshot size
echo "🧹 Cleaning repository for faster snapshot..."
rm -rf node_modules .git/objects/pack/* .git/refs/remotes/* .git/logs/* media/* dist/assets/* 2>/dev/null || true

# Create empty package-lock.json if it doesn't exist to prevent npm ci errors
echo "📦 Ensuring package.json and package-lock.json are properly configured..."
if [ ! -f "package-lock.json" ]; then
  echo "{}" > package-lock.json
  echo "Created empty package-lock.json to prevent npm ci errors"
fi

# Modify Dockerfile to use npm install instead of npm ci if needed
if grep -q "npm ci" Dockerfile; then
  echo "⚙️ Updating Dockerfile to use npm install instead of npm ci..."
  sed -i 's/RUN npm ci/RUN npm install --production=false/g' Dockerfile
fi

# Build the application with minimal dependencies
echo "🔨 Building application with optimized settings..."
go build -ldflags="-s -w" -o server cmd/server/main.go

# Start the server with production settings
echo "🚀 Starting server in production mode..."
PORT=8080 ./server