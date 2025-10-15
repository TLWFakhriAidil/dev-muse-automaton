# ⚡ Required Environment Variables - Quick Reference

## 🚨 CRITICAL: Minimum Required Variables

**These 3 variables are MANDATORY for the app to start:**

```bash
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-actual-anon-key-from-supabase
SUPABASE_SERVICE_KEY=your-actual-service-role-key-from-supabase
```

---

## 📍 Where to Find These Values

### Supabase Dashboard
1. Go to: https://app.supabase.com
2. Select your project
3. Navigate to: **Settings** → **API**
4. Copy the values:

| Variable | Supabase Field | Example |
|----------|---------------|---------|
| `SUPABASE_URL` | Project URL | `https://abcdefgh.supabase.co` |
| `SUPABASE_ANON_KEY` | anon public | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` |
| `SUPABASE_SERVICE_KEY` | service_role | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` |

---

## ✅ Copy-Paste Template for Railway

```bash
# === REQUIRED (Add these first) ===
SUPABASE_URL=
SUPABASE_ANON_KEY=
SUPABASE_SERVICE_KEY=

# === RECOMMENDED (Add for full functionality) ===
APP_ENV=production
JWT_SECRET=
SESSION_SECRET=

# === OPTIONAL (Can add later) ===
REDIS_URL=redis://localhost:6379
WHATSAPP_STORAGE_PATH=./whatsapp_sessions
WHATSAPP_SESSION_DIR=./whatsapp_sessions
WHATSAPP_MAX_DEVICES=10
OPENROUTER_DEFAULT_KEY=
OPENROUTER_TIMEOUT=15
OPENROUTER_MAX_RETRIES=2
MAX_CONCURRENT_USERS=5000
WEBSOCKET_ENABLED=true
CDN_ENABLED=false
CDN_BASE_URL=
```

---

## 🔧 How to Add in Railway

1. Go to Railway dashboard
2. Select your **dev-muse-automaton** project
3. Click **Variables** tab
4. Click **"Raw Editor"** button
5. Paste the template above (with your actual values)
6. Click **Save**
7. Wait for automatic redeploy

---

## 🎯 Quick Test

After adding variables, verify deployment:

```bash
# Check if app is running
curl https://your-railway-app.railway.app/healthz

# Expected response:
{
  "status": "ok",
  "time": 1697395200,
  "database": {
    "status": "connected"
  }
}
```

---

## 🚨 Common Mistakes

❌ **Don't do this:**
- Using example values from `.env.example`
- Forgetting to replace placeholder text
- Adding quotes around values in Railway
- Using anon key instead of service_role key

✅ **Do this:**
- Use actual keys from your Supabase project
- No quotes needed in Railway (just the raw value)
- Use service_role key for `SUPABASE_SERVICE_KEY`
- Double-check URL has no trailing slash

---

**Need detailed setup guide?** See [RAILWAY_DEPLOYMENT_GUIDE.md](./RAILWAY_DEPLOYMENT_GUIDE.md)
