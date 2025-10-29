// Deno Deploy Complete Message Processor
// Purpose:
// 1. Receive webhook messages
// 2. Queue with 30s debouncing (multiple messages = 1 response)
// 3. Process with AI (OpenAI/Anthropic)
// 4. Send response back to WhatsApp

// Environment variables
const DEBOUNCE_DELAY_MS = 30000; // 30 seconds
const OPENAI_API_KEY = Deno.env.get("OPENAI_API_KEY");
const ANTHROPIC_API_KEY = Deno.env.get("ANTHROPIC_API_KEY");
const WHATSAPP_API_URL = Deno.env.get("WHATSAPP_API_URL"); // Your WhatsApp API endpoint

// Open Deno KV database
const kv = await Deno.openKv();

// Message queue structure
interface QueuedMessage {
  phone: string;
  deviceId: string;
  name?: string;
  messages: Array<{
    message: string;
    timestamp: number;
  }>;
  lastMessageTime: number;
  timerScheduled: number;
  conversationHistory?: Array<{ role: string; content: string }>;
}

// Main HTTP handler
Deno.serve(async (req: Request) => {
  const url = new URL(req.url);

  // Health check
  if (url.pathname === "/health") {
    return new Response(
      JSON.stringify({
        status: "ok",
        service: "deno-ai-processor",
        debounceDelay: `${DEBOUNCE_DELAY_MS}ms`
      }),
      { headers: { "Content-Type": "application/json" } }
    );
  }

  // Webhook endpoint - receives messages from WhatsApp
  if (url.pathname === "/webhook" && req.method === "POST") {
    try {
      const payload = await req.json();
      await handleIncomingMessage(payload);

      return new Response(
        JSON.stringify({ success: true, message: "Message queued" }),
        { headers: { "Content-Type": "application/json" } }
      );
    } catch (error) {
      console.error("❌ Webhook error:", error);
      return new Response(
        JSON.stringify({ success: false, error: error.message }),
        { status: 500, headers: { "Content-Type": "application/json" } }
      );
    }
  }

  return new Response("Not Found", { status: 404 });
});

// Handle incoming message
async function handleIncomingMessage(payload: any) {
  const { phone, deviceId, message, name } = payload;

  if (!phone || !deviceId || !message) {
    throw new Error("Missing required fields: phone, deviceId, message");
  }

  const queueKey = ["message_queue", deviceId, phone];
  const now = Date.now();

  // Get existing queue
  const result = await kv.get<QueuedMessage>(queueKey);
  let queue: QueuedMessage;

  if (result.value) {
    // Add to existing queue and reset timer
    queue = result.value;
    queue.messages.push({ message, timestamp: now });
    queue.lastMessageTime = now;
    queue.timerScheduled = now + DEBOUNCE_DELAY_MS;

    console.log(
      `📩 [${phone}] Message ${queue.messages.length} added. Timer RESET to 30s.`
    );
  } else {
    // Create new queue
    queue = {
      phone,
      deviceId,
      name,
      messages: [{ message, timestamp: now }],
      lastMessageTime: now,
      timerScheduled: now + DEBOUNCE_DELAY_MS,
      conversationHistory: [],
    };

    console.log(`🆕 [${phone}] New queue created. Timer started (30s).`);
  }

  // Save queue
  await kv.set(queueKey, queue);

  // Schedule processing
  scheduleProcessing(phone, deviceId, queue.timerScheduled);
}

// Schedule message processing
function scheduleProcessing(phone: string, deviceId: string, scheduledTime: number) {
  const delay = scheduledTime - Date.now();

  if (delay > 0) {
    setTimeout(async () => {
      await checkAndProcess(phone, deviceId, scheduledTime);
    }, delay);
  }
}

// Check timer and process if expired
async function checkAndProcess(
  phone: string,
  deviceId: string,
  scheduledTime: number
) {
  const queueKey = ["message_queue", deviceId, phone];
  const result = await kv.get<QueuedMessage>(queueKey);

  if (!result.value) {
    console.log(`⚠️ [${phone}] Queue not found - already processed`);
    return;
  }

  const queue = result.value;
  const now = Date.now();

  // Check if timer was reset by new message
  if (queue.timerScheduled !== scheduledTime) {
    console.log(`⏭️ [${phone}] Timer was reset - skipping`);
    return;
  }

  // Check if time expired
  if (now >= queue.timerScheduled) {
    console.log(
      `⏰ [${phone}] Timer EXPIRED! Processing ${queue.messages.length} messages...`
    );
    await processMessagesWithAI(queue);
  }
}

// Process messages with AI and send response
async function processMessagesWithAI(queue: QueuedMessage) {
  const { phone, deviceId, messages, conversationHistory } = queue;

  try {
    // Combine all messages
    const combinedMessage = messages.map((m) => m.message).join("\n\n");

    console.log(`🤖 [${phone}] Processing with AI...`);
    console.log(`📝 Combined message:\n${combinedMessage}`);

    // Call AI API (OpenAI/Anthropic)
    const aiResponse = await callAI(combinedMessage, conversationHistory);

    console.log(`✅ [${phone}] AI response: ${aiResponse.substring(0, 100)}...`);

    // Send response via WhatsApp
    await sendWhatsAppMessage(deviceId, phone, aiResponse);

    console.log(`📤 [${phone}] Response sent!`);

    // Update conversation history
    const updatedHistory = [
      ...(conversationHistory || []),
      { role: "user", content: combinedMessage },
      { role: "assistant", content: aiResponse },
    ];

    // Save updated conversation history
    const queueKey = ["message_queue", deviceId, phone];
    await kv.set(queueKey, {
      ...queue,
      conversationHistory: updatedHistory.slice(-10), // Keep last 10 exchanges
      messages: [], // Clear processed messages
    });

    // Delete queue after successful processing
    await kv.delete(queueKey);
    console.log(`🗑️ [${phone}] Queue cleared`);
  } catch (error) {
    console.error(`❌ [${phone}] Processing error:`, error);
    throw error;
  }
}

// Call AI API (OpenAI or Anthropic)
async function callAI(
  message: string,
  history: Array<{ role: string; content: string }> = []
): Promise<string> {
  // Try Anthropic first (Claude)
  if (ANTHROPIC_API_KEY) {
    return await callAnthropic(message, history);
  }

  // Fallback to OpenAI
  if (OPENAI_API_KEY) {
    return await callOpenAI(message, history);
  }

  throw new Error("No AI API key configured");
}

// Call Anthropic Claude API
async function callAnthropic(
  message: string,
  history: Array<{ role: string; content: string }>
): Promise<string> {
  const messages = [
    ...history,
    { role: "user", content: message },
  ];

  const response = await fetch("https://api.anthropic.com/v1/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": ANTHROPIC_API_KEY!,
      "anthropic-version": "2023-06-01",
    },
    body: JSON.stringify({
      model: "claude-3-5-sonnet-20241022",
      max_tokens: 1024,
      messages: messages,
    }),
  });

  if (!response.ok) {
    throw new Error(`Anthropic API error: ${response.status}`);
  }

  const data = await response.json();
  return data.content[0].text;
}

// Call OpenAI GPT API
async function callOpenAI(
  message: string,
  history: Array<{ role: string; content: string }>
): Promise<string> {
  const messages = [
    { role: "system", content: "You are a helpful AI assistant." },
    ...history,
    { role: "user", content: message },
  ];

  const response = await fetch("https://api.openai.com/v1/chat/completions", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${OPENAI_API_KEY}`,
    },
    body: JSON.stringify({
      model: "gpt-4",
      messages: messages,
      max_tokens: 1024,
    }),
  });

  if (!response.ok) {
    throw new Error(`OpenAI API error: ${response.status}`);
  }

  const data = await response.json();
  return data.choices[0].message.content;
}

// Send message via WhatsApp
async function sendWhatsAppMessage(
  deviceId: string,
  phone: string,
  message: string
): Promise<void> {
  if (!WHATSAPP_API_URL) {
    console.warn("⚠️ WHATSAPP_API_URL not configured - skipping send");
    return;
  }

  // Send to your WhatsApp API endpoint
  const response = await fetch(WHATSAPP_API_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      device_id: deviceId,
      phone: phone,
      message: message,
    }),
  });

  if (!response.ok) {
    throw new Error(`WhatsApp API error: ${response.status}`);
  }

  console.log(`✅ WhatsApp message sent to ${phone}`);
}

// Cleanup old queues
async function cleanupOldQueues() {
  const entries = kv.list<QueuedMessage>({ prefix: ["message_queue"] });
  const now = Date.now();
  let cleaned = 0;

  for await (const entry of entries) {
    const queue = entry.value;
    const age = now - queue.lastMessageTime;

    // Delete queues older than 10 minutes
    if (age > 600000) {
      await kv.delete(entry.key);
      cleaned++;
      console.log(`🧹 Cleaned old queue: ${queue.phone} (age: ${age}ms)`);
    }
  }

  if (cleaned > 0) {
    console.log(`🧹 Cleanup: ${cleaned} old queues removed`);
  }
}

// Run cleanup every 10 minutes
setInterval(cleanupOldQueues, 600000);

console.log("🚀 Deno AI Processor with Debouncing Started!");
console.log(`⏱️  Debounce: ${DEBOUNCE_DELAY_MS}ms (30 seconds)`);
console.log(`🤖 AI: ${ANTHROPIC_API_KEY ? "Anthropic Claude" : OPENAI_API_KEY ? "OpenAI GPT" : "NOT CONFIGURED"}`);
console.log(`📱 WhatsApp: ${WHATSAPP_API_URL || "NOT CONFIGURED"}`);
