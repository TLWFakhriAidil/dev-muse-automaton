-- Create chatbot_flows_nodepath table for storing chatbot flows
CREATE TABLE IF NOT EXISTS public.chatbot_flows_nodepath (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  niche TEXT DEFAULT '',
  id_device TEXT DEFAULT '',
  nodes JSONB DEFAULT '[]'::jsonb,
  edges JSONB DEFAULT '[]'::jsonb,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Enable RLS
ALTER TABLE public.chatbot_flows_nodepath ENABLE ROW LEVEL SECURITY;

-- Create policies for authenticated users to manage their own flows
CREATE POLICY "Users can view all flows" 
  ON public.chatbot_flows_nodepath 
  FOR SELECT 
  USING (auth.uid() IS NOT NULL);

CREATE POLICY "Users can insert flows" 
  ON public.chatbot_flows_nodepath 
  FOR INSERT 
  WITH CHECK (auth.uid() IS NOT NULL);

CREATE POLICY "Users can update flows" 
  ON public.chatbot_flows_nodepath 
  FOR UPDATE 
  USING (auth.uid() IS NOT NULL);

CREATE POLICY "Users can delete flows" 
  ON public.chatbot_flows_nodepath 
  FOR DELETE 
  USING (auth.uid() IS NOT NULL);

-- Create trigger for automatic timestamp updates
CREATE TRIGGER update_chatbot_flows_updated_at
  BEFORE UPDATE ON public.chatbot_flows_nodepath
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at_column();