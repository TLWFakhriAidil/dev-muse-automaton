#!/bin/bash
# Database migration script for Supabase
# This script applies the rebuild_schema.sql to your Supabase database

set -e

echo "🔄 Starting Supabase database migration for Chatbot Automation"

# Check if SUPABASE_URL and SUPABASE_KEY are set
if [ -z "$SUPABASE_URL" ] || [ -z "$SUPABASE_DB_PASSWORD" ]; then
    echo "❌ Error: SUPABASE_URL and SUPABASE_DB_PASSWORD must be set"
    echo "Please set these environment variables before running this script"
    exit 1
fi

# Extract project reference from Supabase URL
PROJECT_REF=$(echo $SUPABASE_URL | grep -o '[a-z0-9]*\.supabase\.co' | cut -d'.' -f1)
if [ -z "$PROJECT_REF" ]; then
    echo "❌ Error: Could not extract project reference from SUPABASE_URL"
    exit 1
fi

echo "📊 Using Supabase project: $PROJECT_REF"

# Install psql if not available
if ! command -v psql &> /dev/null; then
    echo "🔧 Installing PostgreSQL client..."
    apt-get update && apt-get install -y postgresql-client || true
fi

# Connection string for Supabase
CONN_STRING="postgres://postgres:$SUPABASE_DB_PASSWORD@db.$PROJECT_REF.supabase.co:5432/postgres?sslmode=require"

echo "🔄 Applying database schema..."
# Apply the schema
psql "$CONN_STRING" -f migrations/rebuild_schema.sql

echo "✅ Database migration completed successfully!"
echo "🔍 Verifying tables..."

# Verify that tables were created
TABLES=$(psql "$CONN_STRING" -c "\dt" -t | wc -l)
echo "📊 Found $TABLES tables in the database"

# List the tables
echo "📋 Tables in the database:"
psql "$CONN_STRING" -c "\dt"

echo "🚀 Database is ready for use!"