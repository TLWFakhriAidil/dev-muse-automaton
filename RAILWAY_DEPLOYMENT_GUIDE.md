# 🚀 Railway Deployment Guide

## Quick Setup: Required Environment Variables

Your application is failing because **Supabase environment variables are missing** in Railway. Follow these steps to fix it.

---

## 📋 Step 1: Add Environment Variables in Railway

1. Go to your Railway project: **dev-muse-automaton**
2. Click on the **Variables** tab (you're already there in the screenshot)
3. Click **"+ New Variable"** button
4. Add each of the following variables one by one

---

## 🔑 Required Environment Variables

### **Supabase Configuration (REQUIRED)**
These are **mandatory** - the app will not start without them:

```bash
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
SUPABASE_SERVICE_KEY=your-service-role-key-here
```

**Where to find these values:**
1. Go to [Supabase Dashboard](https://app.supabase.com)
2. Select your project
3. Go to **Settings** → **API**
4. Copy:
   - **Project URL** → `SUPABASE_URL`
   - **anon/public key** → `SUPABASE_ANON_KEY`
   - **service_role key** → `SUPABASE_SERVICE_KEY`

---

### **Application Configuration**

```bash
# Server Configuration (Railway auto-sets PORT, but you can customize these)
APP_ENV=production

# Redis Configuration
REDIS_URL=redis://localhost:6379

# WhatsApp Configuration
WHATSAPP_STORAGE_PATH=./whatsapp_sessions
WHATSAPP_SESSION_DIR=./whatsapp_sessions
WHATSAPP_MAX_DEVICES=10

# OpenRouter AI Configuration (Optional - for AI features)
OPENROUTER_DEFAULT_KEY=your-openrouter-api-key
OPENROUTER_TIMEOUT=15
OPENROUTER_MAX_RETRIES=2

# Security Configuration (Generate secure random strings)
JWT_SECRET=your-secure-jwt-secret-here
SESSION_SECRET=your-secure-session-secret-here

# Performance Configuration
MAX_CONCURRENT_USERS=5000
WEBSOCKET_ENABLED=true

# CDN Configuration (Optional)
CDN_ENABLED=false
CDN_BASE_URL=
```

---

## 🎯 Priority Variables (Add These First)

To get your app running immediately, add these **3 critical variables**:

1. **SUPABASE_URL**
2. **SUPABASE_ANON_KEY**  
3. **SUPABASE_SERVICE_KEY**

Once these are added, your app will start successfully! ✅

---

## 📝 Step 2: Add Variables in Railway Dashboard

### Method 1: Using Railway UI
1. Click **"+ New Variable"** button
2. Enter variable name (e.g., `SUPABASE_URL`)
3. Enter variable value (e.g., `https://your-project.supabase.co`)
4. Click **Save**
5. Repeat for all required variables

### Method 2: Using Raw Editor
1. Click **"Raw Editor"** button in Railway Variables tab
2. Paste all variables at once in KEY=VALUE format:
```bash
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_ANON_KEY=your-anon-key-here
SUPABASE_SERVICE_KEY=your-service-role-key-here
APP_ENV=production
JWT_SECRET=your-jwt-secret-here
SESSION_SECRET=your-session-secret-here
```
3. Click **Save**

---

## 🔄 Step 3: Redeploy

After adding the environment variables:
1. Railway will **automatically redeploy** your service
2. Or manually trigger a redeploy by clicking **"Deploy"**
3. Check the logs - you should see:
   ```
   ✅ Supabase database initialized successfully
   ✅ Database migrations completed
   ✅ Server starting on port 8080
   ```

---

## 🛠️ Troubleshooting

### Error: "SUPABASE_URL and SUPABASE_SERVICE_KEY are required"
**Solution**: You haven't added the Supabase environment variables yet. Follow Step 1 above.

### Error: "Failed to connect to Supabase"
**Solution**: 
- Check that your Supabase project is active
- Verify the URL format: `https://xxxxx.supabase.co` (no trailing slash)
- Ensure you're using the **service_role key**, not the anon key for `SUPABASE_SERVICE_KEY`

### Error: "Database connection failed"
**Solution**:
- Make sure Supabase database is not paused
- Check that the service_role key has proper permissions
- Verify your Supabase project region matches your Railway region

---

## 🔐 Security Best Practices

1. **Never commit sensitive keys** to Git
2. **Use different keys** for development and production
3. **Generate strong secrets** for JWT_SECRET and SESSION_SECRET:
   ```bash
   # Generate a secure random string (32+ characters)
   openssl rand -base64 32
   ```
4. **Rotate keys regularly** in production

---

## 📊 Post-Deployment Checklist

After deployment succeeds, verify:
- [ ] Application is running (check Railway logs)
- [ ] Database connection works (check `/healthz` endpoint)
- [ ] User registration works (test `/register` page)
- [ ] Login functionality works (test `/login` page)
- [ ] Device settings accessible
- [ ] WhatsApp integration functional

---

## 🆘 Need Help?

If you're still having issues:
1. Check Railway logs for specific error messages
2. Verify all environment variables are set correctly
3. Ensure Supabase project is active and accessible
4. Check that database migrations have run successfully

---

## 📚 Additional Resources

- [Railway Documentation](https://docs.railway.app)
- [Supabase Documentation](https://supabase.com/docs)
- [Environment Variables Best Practices](https://12factor.net/config)

---

**Last Updated**: October 15, 2025  
**Version**: 1.0.0
