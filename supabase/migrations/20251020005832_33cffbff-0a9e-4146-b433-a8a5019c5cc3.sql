-- Create device_setting_nodepath table
CREATE TABLE IF NOT EXISTS public.device_setting_nodepath (
  id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
  device_id TEXT,
  id_device TEXT,
  webhook_id TEXT,
  instance TEXT,
  provider TEXT DEFAULT 'wablas' CHECK (provider IN ('whacenter', 'wablas', 'waha')),
  api_key TEXT,
  api_key_option TEXT DEFAULT 'openai/gpt-4.1' CHECK (api_key_option IN ('openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5')),
  phone_number TEXT,
  id_admin TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Enable RLS
ALTER TABLE public.device_setting_nodepath ENABLE ROW LEVEL SECURITY;

-- Create policies for authenticated users to manage their devices
CREATE POLICY "Users can view all device settings" 
  ON public.device_setting_nodepath 
  FOR SELECT 
  USING (auth.uid() IS NOT NULL);

CREATE POLICY "Users can insert device settings" 
  ON public.device_setting_nodepath 
  FOR INSERT 
  WITH CHECK (auth.uid() IS NOT NULL);

CREATE POLICY "Users can update device settings" 
  ON public.device_setting_nodepath 
  FOR UPDATE 
  USING (auth.uid() IS NOT NULL);

CREATE POLICY "Users can delete device settings" 
  ON public.device_setting_nodepath 
  FOR DELETE 
  USING (auth.uid() IS NOT NULL);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_device_settings_updated_at
  BEFORE UPDATE ON public.device_setting_nodepath
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();