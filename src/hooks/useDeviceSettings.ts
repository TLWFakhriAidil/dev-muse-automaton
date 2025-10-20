import { useState, useEffect } from 'react';
import { supabase } from '@/integrations/supabase/client';
import { useAuth } from '@/contexts/AuthContext';

interface DeviceSettings {
  id: string;
  id_device: string | null;
  device_id: string | null;
  webhook_id: string | null;
  instance: string | null;
  provider: string | null;
  api_key: string | null;
  api_key_option: string | null;
  phone_number: string | null;
  id_admin: string | null;
  created_at: string;
  updated_at: string;
}

export const useDeviceSettings = () => {
  const { user } = useAuth();
  const [devices, setDevices] = useState<DeviceSettings[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchDevices = async () => {
    if (!user) {
      setDevices([]);
      setIsLoading(false);
      return;
    }

    try {
      setIsLoading(true);
      // Use type assertion until Supabase types regenerate
      const { data, error } = await supabase
        .from('device_setting_nodepath' as any)
        .select('*')
        .order('created_at', { ascending: false });

      if (error) throw error;

      setDevices((data || []) as unknown as DeviceSettings[]);
      setError(null);
    } catch (err: any) {
      console.error('Error fetching devices:', err);
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const createDevice = async (device: any) => {
    if (!user) {
      throw new Error('User not authenticated');
    }

    try {
      // Use type assertion until Supabase types regenerate
      const { data, error } = await supabase
        .from('device_setting_nodepath' as any)
        .insert([device])
        .select()
        .single();

      if (error) throw error;

      await fetchDevices();
      return { success: true, data };
    } catch (err: any) {
      console.error('Error creating device:', err);
      return { success: false, error: err.message };
    }
  };

  const updateDevice = async (id: string, updates: any) => {
    if (!user) {
      throw new Error('User not authenticated');
    }

    try {
      // Use type assertion until Supabase types regenerate
      const { data, error } = await supabase
        .from('device_setting_nodepath' as any)
        .update(updates)
        .eq('id', id)
        .select()
        .single();

      if (error) throw error;

      await fetchDevices();
      return { success: true, data };
    } catch (err: any) {
      console.error('Error updating device:', err);
      return { success: false, error: err.message };
    }
  };

  const deleteDevice = async (id: string) => {
    if (!user) {
      throw new Error('User not authenticated');
    }

    try {
      // Use type assertion until Supabase types regenerate
      const { error } = await supabase
        .from('device_setting_nodepath' as any)
        .delete()
        .eq('id', id);

      if (error) throw error;

      await fetchDevices();
      return { success: true };
    } catch (err: any) {
      console.error('Error deleting device:', err);
      return { success: false, error: err.message };
    }
  };

  useEffect(() => {
    fetchDevices();
  }, [user]);

  return {
    devices,
    isLoading,
    error,
    createDevice,
    updateDevice,
    deleteDevice,
    refetch: fetchDevices,
  };
};
