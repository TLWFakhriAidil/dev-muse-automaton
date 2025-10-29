# Deno Deploy AI Message Processor with Debouncing

## Purpose
Handle WhatsApp messages with 30-second debouncing:
- Customer sends message 1 → Start 30s timer
- Customer sends message 2 → Reset timer to 30s
- Customer sends message 3 → Reset timer to 30s
- After 30s of silence → Process all 3 messages together with AI
- Send ONE response

## Features
- ✅ 30-second message debouncing
- ✅ Deno KV for message queue storage
- ✅ AI processing (Anthropic Claude or OpenAI GPT)
- ✅ Conversation history tracking
- ✅ WhatsApp response sending
- ✅ Automatic cleanup of old queues

## Deployment to Deno Deploy

### 1. Install Deno CLI (if not installed)
```bash
# Windows (PowerShell)
irm https://deno.land/install.ps1 | iex

# macOS/Linux
curl -fsSL https://deno.land/x/install/install.sh | sh
```

### 2. Login to Deno Deploy
```bash
deno install --allow-all --name deployctl jsr:@deno/deployctl
deployctl login
```

### 3. Deploy
```bash
cd deno-debouncer
deployctl deploy --project=your-project-name main.ts
```

### 4. Set Environment Variables in Deno Deploy Dashboard
Go to https://dash.deno.com and set:

```
ANTHROPIC_API_KEY=your-anthropic-api-key
# OR
OPENAI_API_KEY=your-openai-api-key

WHATSAPP_API_URL=https://your-go-backend.railway.app/api/whatsapp/send
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes* | Anthropic Claude API key |
| `OPENAI_API_KEY` | Yes* | OpenAI GPT API key |
| `WHATSAPP_API_URL` | Yes | Your WhatsApp send message endpoint |

*At least one AI API key is required

## API Endpoints

### POST /webhook
Receives incoming WhatsApp messages

**Request:**
```json
{
  "phone": "60123456789",
  "deviceId": "device-uuid",
  "message": "Hello!",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Message queued"
}
```

### GET /health
Health check endpoint

**Response:**
```json
{
  "status": "ok",
  "service": "deno-ai-processor",
  "debounceDelay": "30000ms"
}
```

## Flow Diagram

```
Customer sends message 1
    ↓
Deno receives → Queue created → Timer: 30s
    ↓
Customer sends message 2 (10s later)
    ↓
Add to queue → Timer RESET → Timer: 30s
    ↓
Customer sends message 3 (5s later)
    ↓
Add to queue → Timer RESET → Timer: 30s
    ↓
[30 seconds of silence]
    ↓
Timer expires
    ↓
Combine all 3 messages
    ↓
Send to AI (Claude/GPT)
    ↓
Get AI response
    ↓
Send ONE response to WhatsApp
    ↓
Clear queue
```

## Testing Locally

```bash
# Run locally
deno task dev

# Test webhook
curl -X POST http://localhost:8000/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "60123456789",
    "deviceId": "test-device",
    "message": "Hello!"
  }'

# Check health
curl http://localhost:8000/health
```

## Connecting to Go Backend

Update your Go webhook handler to forward messages to Deno Deploy:

```go
// Forward to Deno Deploy instead of processing directly
denoURL := "https://your-project.deno.dev/webhook"

resp, err := http.Post(denoURL, "application/json", bytes.NewBuffer(jsonData))
```

## Logs

View logs in Deno Deploy dashboard:
- https://dash.deno.com/projects/your-project-name/logs

## Monitoring

Monitor your deployment:
- Queue size: Check Deno KV entries
- Processing time: Check logs for timing
- Error rate: Monitor failed requests

## Cost

Deno Deploy pricing:
- Free tier: 100,000 requests/month
- Deno KV: Included in free tier (10GB storage)

## Support

For issues or questions, check:
- Deno Deploy Docs: https://docs.deno.com/deploy/
- Deno KV Docs: https://docs.deno.com/kv/manual
