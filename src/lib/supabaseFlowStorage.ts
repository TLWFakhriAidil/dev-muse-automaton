// Flow storage using Supabase
import { supabase } from '@/integrations/supabase/client';
import { ChatbotFlow } from '@/types/chatbot';

export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  try {
    console.log('Saving flow to Supabase:', flow.id);

    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User not authenticated');
    }

    // Prepare flow data
    const flowData = {
      id: flow.id,
      name: flow.name || 'Untitled Flow',
      niche: flow.niche || '',
      id_device: flow.selectedDeviceId || '',
      nodes: JSON.parse(JSON.stringify(flow.nodes || [])),
      edges: JSON.parse(JSON.stringify(flow.edges || [])),
      updated_at: new Date().toISOString(),
    };

    // Upsert the flow
    const { error } = await supabase
      .from('chatbot_flows_nodepath')
      .upsert([flowData]);

    if (error) {
      console.error('Error saving flow:', error);
      throw error;
    }

    console.log('Flow saved successfully to Supabase');
  } catch (error) {
    console.error('Failed to save flow:', error);
    throw error;
  }
};

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User not authenticated');
    }

    const { data, error } = await supabase
      .from('chatbot_flows_nodepath')
      .select('*')
      .order('updated_at', { ascending: false });

    if (error) {
      console.error('Error fetching flows:', error);
      throw error;
    }

    return (data || []).map(flow => ({
      id: flow.id,
      name: flow.name,
      niche: flow.niche,
      selectedDeviceId: flow.id_device,
      nodes: (Array.isArray(flow.nodes) ? flow.nodes : []) as any,
      edges: (Array.isArray(flow.edges) ? flow.edges : []) as any,
      description: '',
      createdAt: flow.created_at,
      updatedAt: flow.updated_at,
    }));
  } catch (error) {
    console.error('Failed to get flows:', error);
    return [];
  }
};

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User not authenticated');
    }

    const { data, error } = await supabase
      .from('chatbot_flows_nodepath')
      .select('*')
      .eq('id', id)
      .single();

    if (error) {
      console.error('Error fetching flow:', error);
      return null;
    }

    if (!data) return null;

    return {
      id: data.id,
      name: data.name,
      niche: data.niche,
      selectedDeviceId: data.id_device,
      nodes: (Array.isArray(data.nodes) ? data.nodes : []) as any,
      edges: (Array.isArray(data.edges) ? data.edges : []) as any,
      description: '',
      createdAt: data.created_at,
      updatedAt: data.updated_at,
    };
  } catch (error) {
    console.error('Failed to get flow:', error);
    return null;
  }
};

export const deleteFlow = async (id: string): Promise<void> => {
  try {
    const { data: { user } } = await supabase.auth.getUser();
    if (!user) {
      throw new Error('User not authenticated');
    }

    const { error } = await supabase
      .from('chatbot_flows_nodepath')
      .delete()
      .eq('id', id);

    if (error) {
      console.error('Error deleting flow:', error);
      throw error;
    }

    console.log('Flow deleted successfully from Supabase');
  } catch (error) {
    console.error('Failed to delete flow:', error);
    throw error;
  }
};
