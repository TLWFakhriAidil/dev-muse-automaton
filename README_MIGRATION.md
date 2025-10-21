# Chatbot Automation - Migration Complete ✅

## Project Information

**Project Name:** Chatbot Automation
**Go Module:** `chatbot-automation`
**Railway URL:** `chatbot-automation-production.up.railway.app`
**Database:** Supabase PostgreSQL

---

## ⚠️ IMPORTANT: Database Table Names

### Current Table Names (WITHOUT _nodepath suffix):
- ✅ `chatbot_flows`
- ✅ `device_setting`
- ✅ `ai_whatsapp`
- ✅ `wasapBot`
- ✅ `profiles`

### ❌ OLD Table Names (DO NOT USE):
- ❌ `chatbot_flows_nodepath`
- ❌ `device_setting_nodepath`
- ❌ `ai_whatsapp_nodepath`
- ❌ `wasapBot_nodepath`

---

## Migration Completed

### ✅ What Was Changed:

1. **Go Module Name**
   - Changed from `nodepath-chat` to `chatbot-automation`
   - All import paths updated in 67 Go files

2. **Database Table Names**
   - Removed `_nodepath` suffix from all table references
   - Frontend TypeScript types updated
   - All hooks and components updated

3. **Test Email Addresses**
   - `admin@chatbot-automation.com` (password: admin123)
   - `test@chatbot-automation.com` (password: test123)
   - `demo@chatbot-automation.com` (password: demo123)

4. **Branding**
   - UI displays "Chatbot Automation"
   - Copyright footer updated
   - Login/Register pages updated

5. **Railway URLs**
   - All webhooks use `chatbot-automation-production.up.railway.app`
   - Vite config updated
   - OpenRouter API headers updated

---

## 🚀 Deployment Instructions

### Step 1: Run Supabase Migration (REQUIRED!)

Before deploying code changes, you MUST rename the tables in Supabase:

1. Go to Supabase SQL Editor: https://app.supabase.com/project/_/sql
2. Run the migration script: [`SUPABASE_MIGRATION.sql`](./SUPABASE_MIGRATION.sql)
3. Verify all tables renamed successfully

**SQL to run:**
```sql
ALTER TABLE IF EXISTS chatbot_flows_nodepath RENAME TO chatbot_flows;
ALTER TABLE IF EXISTS device_setting_nodepath RENAME TO device_setting;
ALTER TABLE IF EXISTS ai_whatsapp_nodepath RENAME TO ai_whatsapp;
ALTER TABLE IF EXISTS wasapBot_nodepath RENAME TO wasapBot;
```

### Step 2: Deploy to Railway

Once Supabase tables are renamed, deploy the updated code:

```bash
git add .
git commit -m "Complete migration to chatbot-automation"
git push
```

Railway will automatically redeploy with:
- New Go module name
- Updated table references
- New branding

---

## 📝 Code Structure

### Go Module (`chatbot-automation`)

All Go files now import using:
```go
import (
    "chatbot-automation/internal/config"
    "chatbot-automation/internal/models"
    "chatbot-automation/internal/repository"
)
```

### Database Access

All code now references tables WITHOUT `_nodepath` suffix:

```go
// Go code
supabase.From("chatbot_flows").Select("*")
supabase.From("device_setting").Select("*")
supabase.From("ai_whatsapp").Select("*")
```

```typescript
// TypeScript code
supabase.from('chatbot_flows').select('*')
supabase.from('device_setting').select('*')
supabase.from('ai_whatsapp').select('*')
```

---

## ⚙️ Environment Variables

### Railway Environment Variables Required:

**Backend (Go):**
```
SUPABASE_URL=https://bjnjucwpwdzgsnqmpmff.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key
SUPABASE_DB_PASSWORD=your-db-password
PORT=8080
```

**Frontend (Vite Build-Time):**
```
VITE_SUPABASE_URL=https://bjnjucwpwdzgsnqmpmff.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key
```

---

## 🧪 Testing

### Test Credentials

Use these credentials for testing:
- Email: `admin@chatbot-automation.com`
- Password: `admin123`

OR

- Email: `test@chatbot-automation.com`
- Password: `test123`

### Verify Migration

1. **Check Table Names:**
   ```sql
   SELECT table_name
   FROM information_schema.tables
   WHERE table_schema = 'public'
   ORDER BY table_name;
   ```

2. **Test Frontend:**
   - Visit: https://chatbot-automation-production.up.railway.app
   - Should see "Chatbot Automation" branding
   - Login should work
   - Device settings should load

3. **Test Backend:**
   ```bash
   curl https://chatbot-automation-production.up.railway.app/healthz
   ```

---

## 🔧 Development

### Local Development

```bash
# Install dependencies
npm install
go mod download

# Run frontend
npm run dev

# Run backend
go run cmd/server/main.go
```

### Build

```bash
# Frontend
npm run build

# Backend
CGO_ENABLED=0 go build -o app.exe cmd/server/main.go
```

---

## 📚 File Structure

```
chatbot-automation/
├── cmd/server/main.go           # Main entry point
├── internal/
│   ├── config/                   # Configuration
│   ├── database/                 # Database connections
│   ├── handlers/                 # HTTP handlers
│   ├── models/                   # Data models
│   ├── repository/               # Database repositories
│   ├── services/                 # Business logic
│   └── whatsapp/                 # WhatsApp integration
├── src/                          # React frontend
│   ├── components/
│   ├── pages/
│   ├── hooks/
│   └── integrations/supabase/
├── go.mod                        # Go module (chatbot-automation)
├── package.json                  # Node dependencies
├── Dockerfile                    # Railway deployment
├── railway.toml                  # Railway configuration
└── SUPABASE_MIGRATION.sql       # Database migration script
```

---

## ❌ What NOT to Change

### DO NOT Change These:

1. **Module name in go.mod** - Already set to `chatbot-automation`
2. **Table names** - Already correct (no `_nodepath` suffix)
3. **Import paths** - Already using `chatbot-automation/*`

---

## 🔄 Rollback (If Needed)

If you need to revert the migration:

### Supabase Rollback:
```sql
ALTER TABLE IF EXISTS chatbot_flows RENAME TO chatbot_flows_nodepath;
ALTER TABLE IF EXISTS device_setting RENAME TO device_setting_nodepath;
ALTER TABLE IF EXISTS ai_whatsapp RENAME TO ai_whatsapp_nodepath;
ALTER TABLE IF EXISTS wasapBot RENAME TO wasapBot_nodepath;
```

Then revert the code changes using git:
```bash
git revert HEAD
git push
```

---

## 📞 Support

For issues or questions:
1. Check Railway logs: https://railway.app/dashboard
2. Check Supabase logs: https://app.supabase.com/project/_/logs
3. Review this README for migration steps

---

**Last Updated:** October 21, 2025
**Migration Status:** ✅ COMPLETE
