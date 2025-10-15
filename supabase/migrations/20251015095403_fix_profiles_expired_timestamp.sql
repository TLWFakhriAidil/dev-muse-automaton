-- Fix handle_new_user() function to set expired timestamp
-- When a new user registers, set expired to NOW() + 7 days (Trial period)

CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
  INSERT INTO public.profiles (id, full_name, phone, gmail, status, expired)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'full_name', ''),
    COALESCE(NEW.raw_user_meta_data->>'phone', NULL),
    COALESCE(NEW.raw_user_meta_data->>'gmail', NULL),
    'Trial',
    NOW() + INTERVAL '7 days'  -- Set expired to current timestamp + 7 days for Trial users
  );
  RETURN NEW;
END;
$$;

-- Update existing NULL expired timestamps to NOW() + 7 days
UPDATE public.profiles
SET expired = NOW() + INTERVAL '7 days'
WHERE expired IS NULL AND status = 'Trial';

COMMENT ON FUNCTION public.handle_new_user() IS 'Automatically creates a profile for new users with a 7-day trial period';
