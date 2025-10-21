-- ============================================================================
-- SUPABASE TABLE MIGRATION: Remove _nodepath suffix
-- ============================================================================
--
-- IMPORTANT: Run this BEFORE deploying the updated code!
--
-- This script renames all tables to remove the _nodepath suffix
-- Run this in your Supabase SQL Editor: https://app.supabase.com/project/_/sql
--
-- ============================================================================

-- Step 1: Rename all tables
-- ============================================================================

-- Rename chatbot flows table
ALTER TABLE IF EXISTS chatbot_flows_nodepath RENAME TO chatbot_flows;

-- Rename device settings table
ALTER TABLE IF EXISTS device_setting_nodepath RENAME TO device_setting;

-- Rename AI WhatsApp conversations table
ALTER TABLE IF EXISTS ai_whatsapp_nodepath RENAME TO ai_whatsapp;

-- Rename WasapBot table
ALTER TABLE IF EXISTS wasapBot_nodepath RENAME TO wasapBot;

-- Step 2: Verify the migration
-- ============================================================================

-- Check that old tables no longer exist
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'chatbot_flows_nodepath') THEN
        RAISE EXCEPTION 'Migration failed: chatbot_flows_nodepath still exists';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'device_setting_nodepath') THEN
        RAISE EXCEPTION 'Migration failed: device_setting_nodepath still exists';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_whatsapp_nodepath') THEN
        RAISE EXCEPTION 'Migration failed: ai_whatsapp_nodepath still exists';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'wasapBot_nodepath') THEN
        RAISE EXCEPTION 'Migration failed: wasapBot_nodepath still exists';
    END IF;

    RAISE NOTICE 'Migration successful! All tables renamed.';
END $$;

-- Step 3: List all tables to confirm
-- ============================================================================

SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- ============================================================================
-- ROLLBACK SCRIPT (if you need to undo this migration)
-- ============================================================================
--
-- ALTER TABLE IF EXISTS chatbot_flows RENAME TO chatbot_flows_nodepath;
-- ALTER TABLE IF EXISTS device_setting RENAME TO device_setting_nodepath;
-- ALTER TABLE IF EXISTS ai_whatsapp RENAME TO ai_whatsapp_nodepath;
-- ALTER TABLE IF EXISTS wasapBot RENAME TO wasapBot_nodepath;
--
-- ============================================================================
