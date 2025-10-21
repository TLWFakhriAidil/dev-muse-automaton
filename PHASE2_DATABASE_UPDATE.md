# Phase 2: Database Schema Update

## Overview
This document guides you through executing the **FINAL** database schema that matches your existing MySQL structure.

## What Changed

### New Schema (supabase_schema_final.sql)
- **10 tables** matching your existing system:
  1. `user` - User registration and profile
  2. `user_sessions` - Session login logs
  3. `device_setting` - Device registration (WhatsApp providers)
  4. `chatbot_flows` - User-created flow nodes (JSONB)
  5. `ai_whatsapp` - Chatbot AI incoming messages
  6. `ai_whatsapp_session` - Session lockout for AI
  7. `wasapBot` - WasapBot Exama incoming messages
  8. `wasapBot_session` - Session lockout for WasapBot
  9. `stageSetValue` - Stage configuration for WasapBot
  10. `orders` - Billplz payment billing records

### Key Features
- ✅ Auto-updating timestamps (PostgreSQL triggers)
- ✅ Row Level Security (RLS) for user data isolation
- ✅ JSONB for flows and conversation history
- ✅ Session locking to prevent duplicate message processing
- ✅ Billplz payment integration table
- ✅ Multiple WhatsApp provider support (Waha, Wablas, Whacenter)
- ✅ AI model options (GPT-4.1, GPT-5, Gemini 2.5 Pro, etc.)

## Execute Schema in Supabase

### Step 1: Open Supabase SQL Editor
1. Go to: https://app.supabase.com/project/bjnjucwpwdzgsnqmpmff/sql
2. Click **"New query"**

### Step 2: Copy Schema
1. Open `supabase_schema_final.sql` from the repository
2. Copy **ALL 345 lines**

### Step 3: Execute
1. Paste into SQL Editor
2. Click **"Run"** (or press Ctrl+Enter)
3. Wait for execution (should take ~5-10 seconds)

### Step 4: Verify Tables Created
Run this query to verify all 10 tables exist:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
ORDER BY table_name;
```

**Expected output:**
```
ai_whatsapp
ai_whatsapp_session
chatbot_flows
device_setting
orders
stageSetValue
user
user_sessions
wasapBot
wasapBot_session
```

### Step 5: Verify Test User Created
Run this query:

```sql
SELECT id, email, full_name, status, created_at
FROM "user"
LIMIT 1;
```

**Expected output:**
```
email: test@chatbot-automation.com
full_name: Test User
status: Trial
```

## Backend Code Updated

The following files have been updated to match the new schema:

### 1. internal/database/supabase.go
- Changed `TestConnection()` from querying `users` → `user` table

### 2. cmd/server/main.go
- Changed `/api/db/test` endpoint from querying `users` → `user` table

## After Schema Execution

Once you've executed the schema in Supabase, verify the backend connection:

### Test 1: Connection Status
```
https://chatbot-automation-production.up.railway.app/api/status
```

**Expected response:**
```json
{
  "status": "running",
  "version": "2.0.0-rebuild",
  "database": "connected"
}
```

### Test 2: Database Query Test
```
https://chatbot-automation-production.up.railway.app/api/db/test
```

**Expected response:**
```json
{
  "success": true,
  "message": "Database connection working!",
  "users": "[{\"id\":\"...\",\"email\":\"test@chatbot-automation.com\",\"full_name\":\"Test User\",\"status\":\"Trial\"}]"
}
```

## Important Notes

### Table Name with Quotes
PostgreSQL requires quotes around the `user` table name because "user" is a reserved keyword:
```sql
SELECT * FROM "user";  -- ✅ Correct
SELECT * FROM user;    -- ❌ Error: syntax error
```

### Row Level Security (RLS)
- All user-facing tables have RLS enabled
- Users can only see their own data
- Backend uses `service_role` key for full access
- Conversation tables (ai_whatsapp, wasapBot) only accessible via service role

### Auto-Updating Timestamps
These tables automatically update `updated_at` on every UPDATE:
- user
- device_setting
- chatbot_flows
- ai_whatsapp
- orders

## Next Steps After Schema Execution

1. ✅ Execute schema in Supabase (this step)
2. 🔄 Verify connection with `/api/status` and `/api/db/test`
3. ⏳ Phase 3: Build authentication system (login/register endpoints)
4. ⏳ Phase 4: Device management UI
5. ⏳ Phase 5: Flow builder with React Flow
6. ⏳ Phase 6: WhatsApp webhook integration
7. ⏳ Phase 7: AI conversation engine

## Troubleshooting

### If you get "relation does not exist" error:
The schema hasn't been executed yet. Follow Steps 1-3 above.

### If you get "permission denied" error:
You're using the anon key. Backend should use service_role key for these tables.

### If timestamps don't auto-update:
The triggers are automatically created by the schema. Verify by running:
```sql
SELECT tgname FROM pg_trigger WHERE tgname LIKE 'update_%';
```

## Schema Differences from Previous Version

| Feature | Old Schema | New Schema |
|---------|-----------|------------|
| Tables | 9 tables | 10 tables (added `orders`) |
| Table name | `users` | `user` (matches your MySQL) |
| Billplz | ❌ Not included | ✅ Full `orders` table |
| Session locking | ✅ ai_whatsapp_session only | ✅ Both AI and WasapBot sessions |
| WasapBot tables | ❌ Missing | ✅ `wasapBot` + `wasapBot_session` + `stageSetValue` |

---

**Status:** Ready to execute
**Estimated time:** 30 seconds to copy/paste and run
**File to execute:** `supabase_schema_final.sql` (345 lines)
