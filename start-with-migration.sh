#!/bin/bash

# Railway startup script with database migration
# This script runs the migration before starting the main server

echo "🚀 Starting Railway deployment with database migration..."

# Check if SUPABASE_URL is available for PostgreSQL migration
if [ -n "$SUPABASE_URL" ] && [ -n "$SUPABASE_DB_PASSWORD" ]; then
    echo "📊 SUPABASE_URL found, running PostgreSQL migration..."
    
    # Check if the migration script exists
    if [ -f "/app/railway-db-migration.sh" ]; then
        echo "🔄 Executing Supabase database migration..."
        chmod +x /app/railway-db-migration.sh
        /app/railway-db-migration.sh
        
        supabase_migration_exit_code=$?
        if [ $supabase_migration_exit_code -eq 0 ]; then
            echo "✅ Supabase migration completed successfully"
        else
            echo "⚠️ Supabase migration failed with exit code $supabase_migration_exit_code, but continuing..."
        fi
    else
        echo "⚠️ Supabase migration script not found, skipping"
    fi
else
    echo "⚠️ SUPABASE_URL or SUPABASE_DB_PASSWORD not found, skipping Supabase migration"
fi

# Legacy MySQL migration support
if [ -n "$MYSQL_URI" ]; then
    echo "📡 MYSQL_URI found, running comprehensive migration..."
    
    # Run the Railway migration runner
    echo "🔄 Executing comprehensive database migration..."
    /app/railway_migration_runner
    
    migration_exit_code=$?
    if [ $migration_exit_code -eq 0 ]; then
        echo "✅ Comprehensive migration completed successfully"
    else
        echo "⚠️ Migration failed with exit code $migration_exit_code, but continuing..."
        echo "ℹ️ Server will start anyway to maintain service availability"
    fi
fi

echo "🚀 Starting main server..."
# Start the main application
exec /app/server