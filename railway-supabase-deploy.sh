#!/bin/bash
# railway-supabase-deploy.sh
# Optimized deployment script for Railways with Supabase connection

set -e  # Exit on any error

echo "🚀 Starting Railways deployment with Supabase optimization..."

# Verify required environment variables
if [ -z "$SUPABASE_URL" ]; then
  echo "❌ ERROR: SUPABASE_URL environment variable is required"
  exit 1
fi

if [ -z "$SUPABASE_DB_PASSWORD" ]; then
  echo "❌ ERROR: SUPABASE_DB_PASSWORD environment variable is required"
  exit 1
fi

if [ -z "$SUPABASE_SERVICE_KEY" ]; then
  echo "❌ ERROR: SUPABASE_SERVICE_KEY environment variable is required"
  exit 1
fi

echo "✅ Environment variables verified"

# Build the application
echo "🔨 Building application..."
go build -o server cmd/server/main.go

# Verify build success
if [ ! -f "./server" ]; then
  echo "❌ ERROR: Build failed - server binary not created"
  exit 1
fi

echo "✅ Build successful"

# Test Supabase connection
echo "🔌 Testing Supabase connection..."
go run test_supabase.go

# Start the server
echo "🚀 Starting server..."
./server