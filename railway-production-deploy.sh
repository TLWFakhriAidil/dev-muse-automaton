#!/bin/bash
# Production deployment script for Railway with enhanced Supabase connection handling

# Exit on error
set -e

echo "🚀 Starting production deployment for Chatbot Automation on Railway"

# Verify required environment variables
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_DB_PASSWORD" ]; then
  echo "❌ ERROR: SUPABASE_URL and SUPABASE_DB_PASSWORD environment variables are required"
  exit 1
fi

echo "✅ Environment variables verified"

# Build the application
echo "🔨 Building application..."
go build -o server cmd/server/main.go

# Test Supabase connection
echo "🔌 Testing Supabase connection..."
go run test_supabase.go

# If we get here, the connection test passed
echo "✅ Supabase connection test passed"

# Start the server with production settings
echo "🚀 Starting server in production mode..."
PORT=8080 ./server

# This script should be executed by Railway's deployment process