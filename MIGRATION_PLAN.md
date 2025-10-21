# Complete Migration Plan: NodePath → Chatbot Automation

## Overview
This document outlines the complete migration from "NodePath" branding to "Chatbot Automation" including:
1. Go module renaming
2. Database table name changes (removing `_nodepath` suffix)
3. URL updates
4. Email domain changes

## ⚠️ CRITICAL: DO THIS IN ORDER

### Phase 1: Database Table Renaming (Supabase)

**Current Tables (WITH `_nodepath` suffix):**
- `chatbot_flows_nodepath`
- `device_setting_nodepath`
- `ai_whatsapp_nodepath`
- `wasapBot_nodepath`

**New Tables (WITHOUT suffix):**
- `chatbot_flows`
- `device_setting`
- `ai_whatsapp`
- `wasapBot`

**Supabase SQL Migration:**
```sql
-- Rename tables in Supabase
ALTER TABLE chatbot_flows_nodepath RENAME TO chatbot_flows;
ALTER TABLE device_setting_nodepath RENAME TO device_setting;
ALTER TABLE ai_whatsapp_nodepath RENAME TO ai_whatsapp;
ALTER TABLE wasapBot_nodepath RENAME TO wasapBot;
```

### Phase 2: Go Module Renaming

**Current:** `module nodepath-chat`
**New:** `module chatbot-automation`

**Steps:**
1. Update `go.mod`: Change module name
2. Run: `find . -name "*.go" -exec sed -i 's/nodepath-chat/chatbot-automation/g' {} +`
3. Update all import statements

### Phase 3: Frontend Table References

**Update these files:**
- `src/integrations/supabase/types.ts`
- `src/lib/supabaseFlowStorage.ts`
- `src/hooks/useDeviceSettings.ts`
- `src/pages/Dashboard.tsx`
- `src/pages/WhatsAppBot.tsx`

### Phase 4: Test Email Domains

**Current:**
- `admin@nodepath.com`
- `test@nodepath.com`
- `demo@nodepath.com`

**New:**
- `admin@chatbot-automation.com`
- `test@chatbot-automation.com`
- `demo@chatbot-automation.com`

## Files to Update

### Go Files (All *.go)
- Update imports from `nodepath-chat/*` to `chatbot-automation/*`

### TypeScript/React Files
- Update table names from `*_nodepath` to just the base name

### Configuration Files
- `go.mod` - Module name
- Railway environment variables (if any)

## Testing Checklist

- [ ] Supabase tables renamed successfully
- [ ] Go build compiles without errors
- [ ] Frontend connects to correct table names
- [ ] All API endpoints work
- [ ] Authentication works with new email domains
- [ ] Railway deployment succeeds

## Rollback Plan

If something breaks:
1. **Database**: Rename tables back to `*_nodepath` format
2. **Go**: Revert go.mod and imports
3. **Frontend**: Revert table name references

## Notes

- **DO NOT** change table names in code before changing them in Supabase
- **DO NOT** commit half-completed migration
- **DO** test locally before deploying to Railway
- **DO** backup Supabase database before table renaming
