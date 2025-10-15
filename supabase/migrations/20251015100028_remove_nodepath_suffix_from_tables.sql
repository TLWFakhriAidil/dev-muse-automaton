-- Remove _nodepath suffix from all table names for cleaner naming
-- This migration renames all tables to remove the _nodepath suffix

-- Rename ai_whatsapp_nodepath to ai_whatsapp
ALTER TABLE IF EXISTS public.ai_whatsapp_nodepath RENAME TO ai_whatsapp;

-- Rename ai_whatsapp_session_nodepath to ai_whatsapp_session
ALTER TABLE IF EXISTS public.ai_whatsapp_session_nodepath RENAME TO ai_whatsapp_session;

-- Rename chatbot_flows_nodepath to chatbot_flows
ALTER TABLE IF EXISTS public.chatbot_flows_nodepath RENAME TO chatbot_flows;

-- Rename device_setting_nodepath to device_setting
ALTER TABLE IF EXISTS public.device_setting_nodepath RENAME TO device_setting;

-- Rename orders_nodepath to orders
ALTER TABLE IF EXISTS public.orders_nodepath RENAME TO orders;

-- Rename stagesetvalue_nodepath to stagesetvalue
ALTER TABLE IF EXISTS public.stagesetvalue_nodepath RENAME TO stagesetvalue;

-- Rename wasapbot_nodepath to wasapbot
ALTER TABLE IF EXISTS public.wasapbot_nodepath RENAME TO wasapbot;

-- Rename wasapbot_session_nodepath to wasapbot_session
ALTER TABLE IF EXISTS public.wasapbot_session_nodepath RENAME TO wasapbot_session;

-- Note: 'profiles' table already has no suffix and doesn't need renaming

-- Add comments to document the change
COMMENT ON TABLE public.ai_whatsapp IS 'AI WhatsApp data table (renamed from ai_whatsapp_nodepath)';
COMMENT ON TABLE public.ai_whatsapp_session IS 'AI WhatsApp session data (renamed from ai_whatsapp_session_nodepath)';
COMMENT ON TABLE public.chatbot_flows IS 'Chatbot flow configurations (renamed from chatbot_flows_nodepath)';
COMMENT ON TABLE public.device_setting IS 'Device settings and configurations (renamed from device_setting_nodepath)';
COMMENT ON TABLE public.orders IS 'Order records (renamed from orders_nodepath)';
COMMENT ON TABLE public.stagesetvalue IS 'Stage set values (renamed from stagesetvalue_nodepath)';
COMMENT ON TABLE public.wasapbot IS 'WasapBot data (renamed from wasapbot_nodepath)';
COMMENT ON TABLE public.wasapbot_session IS 'WasapBot session data (renamed from wasapbot_session_nodepath)';
