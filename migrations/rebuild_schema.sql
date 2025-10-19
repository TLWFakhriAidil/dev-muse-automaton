-- Rebuild schema for Supabase database
-- Created based on system requirements

-- Create profiles table
CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id),
    full_name TEXT,
    phone TEXT,
    gmail TEXT,
    status TEXT,
    expired TIMESTAMPTZ,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create device_settings table for WhatsApp automation
CREATE TABLE IF NOT EXISTS device_settings (
    id_device SERIAL PRIMARY KEY,
    id_user UUID REFERENCES profiles(id),
    device_name TEXT,
    device_type TEXT,
    device_status TEXT DEFAULT 'inactive',
    device_url TEXT,
    device_token TEXT,
    device_setting_nodepath TEXT,
    api_key_option TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create wasapbot table for WhatsApp conversations
CREATE TABLE IF NOT EXISTS wasapbot (
    id_prospect SERIAL PRIMARY KEY,
    id_device INTEGER REFERENCES device_settings(id_device),
    prospect_num TEXT,
    prospect_nama TEXT,
    stage TEXT,
    human INTEGER DEFAULT 0,
    balas TIMESTAMPTZ DEFAULT NOW(),
    conv_last TEXT,
    conv_current TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create ai_settings table for AI configuration
CREATE TABLE IF NOT EXISTS ai_settings (
    id_setting SERIAL PRIMARY KEY,
    id_device INTEGER REFERENCES device_settings(id_device),
    setting_name TEXT,
    setting_value TEXT,
    setting_type TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create stage_set_value table for conversation stages
CREATE TABLE IF NOT EXISTS stage_set_value (
    id_stage SERIAL PRIMARY KEY,
    id_device INTEGER REFERENCES device_settings(id_device),
    stage_name TEXT,
    stage_value TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create execution_process table for tracking processes
CREATE TABLE IF NOT EXISTS execution_process (
    id_process SERIAL PRIMARY KEY,
    id_device INTEGER REFERENCES device_settings(id_device),
    process_type TEXT,
    process_status TEXT,
    process_data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create orders table for billing
CREATE TABLE IF NOT EXISTS orders (
    id_order SERIAL PRIMARY KEY,
    id_user UUID REFERENCES profiles(id),
    order_status TEXT,
    order_amount DECIMAL,
    payment_id TEXT,
    payment_status TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create RLS policies
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE device_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE wasapbot ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE stage_set_value ENABLE ROW LEVEL SECURITY;
ALTER TABLE execution_process ENABLE ROW LEVEL SECURITY;
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Create policies for profiles
CREATE POLICY "Users can view their own profile" 
    ON profiles FOR SELECT 
    USING (auth.uid() = id);

CREATE POLICY "Users can update their own profile" 
    ON profiles FOR UPDATE 
    USING (auth.uid() = id);

-- Create policies for device_settings
CREATE POLICY "Users can view their own devices" 
    ON device_settings FOR SELECT 
    USING (auth.uid() = id_user);

CREATE POLICY "Users can update their own devices" 
    ON device_settings FOR UPDATE 
    USING (auth.uid() = id_user);

CREATE POLICY "Users can insert their own devices" 
    ON device_settings FOR INSERT 
    WITH CHECK (auth.uid() = id_user);

-- Create policies for wasapbot
CREATE POLICY "Users can view their own prospects" 
    ON wasapbot FOR SELECT 
    USING (EXISTS (
        SELECT 1 FROM device_settings 
        WHERE device_settings.id_device = wasapbot.id_device 
        AND device_settings.id_user = auth.uid()
    ));

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_device_settings_id_user ON device_settings(id_user);
CREATE INDEX IF NOT EXISTS idx_wasapbot_id_device ON wasapbot(id_device);
CREATE INDEX IF NOT EXISTS idx_wasapbot_prospect_num ON wasapbot(prospect_num);
CREATE INDEX IF NOT EXISTS idx_ai_settings_id_device ON ai_settings(id_device);
CREATE INDEX IF NOT EXISTS idx_stage_set_value_id_device ON stage_set_value(id_device);
CREATE INDEX IF NOT EXISTS idx_execution_process_id_device ON execution_process(id_device);
CREATE INDEX IF NOT EXISTS idx_orders_id_user ON orders(id_user);