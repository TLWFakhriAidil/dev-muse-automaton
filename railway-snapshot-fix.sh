#!/bin/bash
# Optimized deployment script for Railway with snapshot timeout fix

# Exit on error
set -e

echo "🚀 Starting optimized deployment for Chatbot Automation on Railway"

# Clean up unnecessary files to reduce snapshot size
echo "🧹 Cleaning repository for faster snapshot..."
rm -rf node_modules .git/objects/pack/* .git/refs/remotes/* .git/logs/* media/* dist/assets/* 2>/dev/null || true

# Build the application with minimal dependencies
echo "🔨 Building application with optimized settings..."
go build -ldflags="-s -w" -o server cmd/server/main.go

# Start the server with production settings
echo "🚀 Starting server in production mode..."
PORT=8080 ./server