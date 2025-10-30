// Deno Deploy Edge Function - Message Debouncing with Duplicate Prevention
// This function receives WhatsApp messages, queues them, and sends combined messages to the backend

interface Message {
  device_id: string;
  phone: string;
  name?: string;
  message: string;
  timestamp: number;
}

interface Session {
  messages: Message[];
  timer: number | null;
  isProcessing: boolean; // NEW: Flag to prevent duplicate processing
  lastProcessedAt: number | null; // NEW: Track when last batch was processed
}

// In-memory storage for message queues
const messageQueue = new Map<string, Session>();

// Configuration
const DEBOUNCE_DELAY = 8000; // 8 seconds (configurable)
const PROCESSING_COOLDOWN = 30000; // 30 seconds cooldown after processing
const BACKEND_URL = Deno.env.get("BACKEND_URL") || "https://chatbot-automation-production.up.railway.app";
const BACKEND_ENDPOINT = "/api/debounce/process";

// Logging helper with timestamps
function log(level: string, message: string, data?: any) {
  const timestamp = new Date().toISOString();
  const logEntry = {
    timestamp,
    level,
    message,
    ...(data && { data }),
  };
  console.log(JSON.stringify(logEntry));
}

// Generate unique session key for each device + phone combination
function getSessionKey(deviceId: string, phone: string): string {
  return `${deviceId}:${phone}`;
}

// Check if session is still in cooldown period
function isInCooldown(session: Session): boolean {
  if (!session.lastProcessedAt) return false;
  const timeSinceProcessing = Date.now() - session.lastProcessedAt;
  return timeSinceProcessing < PROCESSING_COOLDOWN;
}

// Process and send combined messages to backend
async function processMessages(sessionKey: string) {
  const session = messageQueue.get(sessionKey);

  if (!session || session.messages.length === 0) {
    log("warn", "No messages to process", { sessionKey });
    return;
  }

  // Check if already processing (prevent duplicate)
  if (session.isProcessing) {
    log("warn", "Session already processing, ignoring duplicate trigger", {
      sessionKey,
      messageCount: session.messages.length
    });
    return;
  }

  // Mark as processing
  session.isProcessing = true;

  const messages = [...session.messages];
  const firstMessage = messages[0];

  log("info", "Processing combined messages", {
    sessionKey,
    messageCount: messages.length,
    deviceId: firstMessage.device_id,
    phone: firstMessage.phone,
    debounceDelay: DEBOUNCE_DELAY,
  });

  try {
    // Prepare payload
    const payload = {
      device_id: firstMessage.device_id,
      phone: firstMessage.phone,
      name: firstMessage.name || "",
      messages: messages.map((m) => m.message),
    };

    log("info", "Sending to backend", {
      url: BACKEND_URL + BACKEND_ENDPOINT,
      payload,
    });

    // Send to backend
    const response = await fetch(BACKEND_URL + BACKEND_ENDPOINT, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    const result = await response.json();

    if (response.ok && result.success) {
      log("success", "Backend processed successfully", {
        sessionKey,
        messageCount: messages.length,
        result,
      });

      // Mark session as successfully processed
      session.lastProcessedAt = Date.now();
      session.messages = []; // Clear processed messages
      session.isProcessing = false; // Reset processing flag

      // Keep session in map for cooldown period, then clean up
      setTimeout(() => {
        messageQueue.delete(sessionKey);
        log("info", "Session cleaned up after cooldown", { sessionKey });
      }, PROCESSING_COOLDOWN);

    } else {
      log("error", "Backend processing failed", {
        sessionKey,
        status: response.status,
        result,
      });

      // Reset processing flag on error so it can retry
      session.isProcessing = false;
    }
  } catch (error) {
    log("error", "Failed to send to backend", {
      sessionKey,
      error: error.message,
    });

    // Reset processing flag on error
    session.isProcessing = false;
  }
}

// Add message to queue and trigger debounce
function queueMessage(msg: Message) {
  const sessionKey = getSessionKey(msg.device_id, msg.phone);

  // Get or create session
  let session = messageQueue.get(sessionKey);

  if (!session) {
    session = {
      messages: [],
      timer: null,
      isProcessing: false,
      lastProcessedAt: null,
    };
    messageQueue.set(sessionKey, session);
    log("info", "New session created", { sessionKey });
  }

  // Check if session is processing or in cooldown
  if (session.isProcessing) {
    log("warn", "Session is processing, message ignored to prevent duplicate", {
      sessionKey,
      ignoredMessage: msg.message.substring(0, 50),
      currentQueueSize: session.messages.length,
    });
    return {
      queued: false,
      reason: "processing",
      message: "Previous batch is being processed. Message ignored to prevent duplicate reply."
    };
  }

  if (isInCooldown(session)) {
    const cooldownRemaining = Math.ceil((PROCESSING_COOLDOWN - (Date.now() - session.lastProcessedAt!)) / 1000);
    log("warn", "Session in cooldown, message ignored", {
      sessionKey,
      cooldownRemaining: `${cooldownRemaining}s`,
      ignoredMessage: msg.message.substring(0, 50),
    });
    return {
      queued: false,
      reason: "cooldown",
      message: `Session in cooldown. Wait ${cooldownRemaining} seconds before sending new messages.`
    };
  }

  // Add message to queue
  session.messages.push(msg);

  log("info", "Message queued", {
    sessionKey,
    queueSize: session.messages.length,
    message: msg.message.substring(0, 50) + "...",
  });

  // Clear existing timer
  if (session.timer !== null) {
    clearTimeout(session.timer);
  }

  // Set new timer
  session.timer = setTimeout(() => {
    processMessages(sessionKey);
  }, DEBOUNCE_DELAY);

  return {
    queued: true,
    queueSize: session.messages.length,
    willProcessIn: DEBOUNCE_DELAY
  };
}

// Main request handler
async function handleRequest(req: Request): Promise<Response> {
  const url = new URL(req.url);

  // Health check endpoint
  if (url.pathname === "/health" || url.pathname === "/") {
    return new Response(
      JSON.stringify({
        status: "ok",
        service: "Deno Deploy - Message Debouncing",
        config: {
          debounceDelay: DEBOUNCE_DELAY,
          processingCooldown: PROCESSING_COOLDOWN,
          backendUrl: BACKEND_URL,
        },
        activeSessions: messageQueue.size,
        timestamp: new Date().toISOString(),
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
  }

  // Status endpoint - show active queues
  if (url.pathname === "/status") {
    const sessions = Array.from(messageQueue.entries()).map(([key, session]) => ({
      sessionKey: key,
      messageCount: session.messages.length,
      isProcessing: session.isProcessing,
      lastProcessedAt: session.lastProcessedAt
        ? new Date(session.lastProcessedAt).toISOString()
        : null,
      inCooldown: isInCooldown(session),
    }));

    return new Response(
      JSON.stringify({
        activeSessions: messageQueue.size,
        sessions,
        timestamp: new Date().toISOString(),
      }),
      {
        headers: { "Content-Type": "application/json" },
      }
    );
  }

  // Queue message endpoint
  if (url.pathname === "/queue" && req.method === "POST") {
    try {
      const body = await req.json();

      // Validate required fields
      if (!body.device_id || !body.phone || !body.message) {
        return new Response(
          JSON.stringify({
            success: false,
            error: "device_id, phone, and message are required",
          }),
          {
            status: 400,
            headers: { "Content-Type": "application/json" },
          }
        );
      }

      const msg: Message = {
        device_id: body.device_id,
        phone: body.phone,
        name: body.name,
        message: body.message,
        timestamp: Date.now(),
      };

      const result = queueMessage(msg);

      return new Response(
        JSON.stringify({
          success: result.queued,
          ...result,
          timestamp: new Date().toISOString(),
        }),
        {
          headers: { "Content-Type": "application/json" },
        }
      );
    } catch (error) {
      log("error", "Failed to queue message", { error: error.message });

      return new Response(
        JSON.stringify({
          success: false,
          error: "Invalid request body",
        }),
        {
          status: 400,
          headers: { "Content-Type": "application/json" },
        }
      );
    }
  }

  // 404 for unknown paths
  return new Response("Not Found", { status: 404 });
}

// Start the server
Deno.serve(handleRequest);

log("info", "Deno Deploy Debounce Service Started", {
  debounceDelay: DEBOUNCE_DELAY,
  processingCooldown: PROCESSING_COOLDOWN,
  backendUrl: BACKEND_URL,
});
