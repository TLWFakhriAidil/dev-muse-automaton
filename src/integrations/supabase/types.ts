export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  // Allows to automatically instantiate createClient with right options
  // instead of createClient<Database, { PostgrestVersion: 'XX' }>(URL, KEY)
  __InternalSupabase: {
    PostgrestVersion: "13.0.5"
  }
  public: {
    Tables: {
      ai_whatsapp_nodepath: {
        Row: {
          balas: string | null
          conv_current: string | null
          conv_last: string | null
          created_at: string
          current_node_id: string | null
          date_order: string | null
          execution_id: string | null
          execution_status: string | null
          flow_id: string | null
          flow_reference: string | null
          human: number | null
          id_device: string | null
          id_prospect: number
          intro: string | null
          keywordiklan: string | null
          last_node_id: string | null
          marketer: string | null
          niche: string | null
          prospect_name: string | null
          prospect_num: string | null
          stage: string | null
          update_today: string | null
          updated_at: string
          waiting_for_reply: boolean | null
        }
        Insert: {
          balas?: string | null
          conv_current?: string | null
          conv_last?: string | null
          created_at?: string
          current_node_id?: string | null
          date_order?: string | null
          execution_id?: string | null
          execution_status?: string | null
          flow_id?: string | null
          flow_reference?: string | null
          human?: number | null
          id_device?: string | null
          id_prospect?: number
          intro?: string | null
          keywordiklan?: string | null
          last_node_id?: string | null
          marketer?: string | null
          niche?: string | null
          prospect_name?: string | null
          prospect_num?: string | null
          stage?: string | null
          update_today?: string | null
          updated_at?: string
          waiting_for_reply?: boolean | null
        }
        Update: {
          balas?: string | null
          conv_current?: string | null
          conv_last?: string | null
          created_at?: string
          current_node_id?: string | null
          date_order?: string | null
          execution_id?: string | null
          execution_status?: string | null
          flow_id?: string | null
          flow_reference?: string | null
          human?: number | null
          id_device?: string | null
          id_prospect?: number
          intro?: string | null
          keywordiklan?: string | null
          last_node_id?: string | null
          marketer?: string | null
          niche?: string | null
          prospect_name?: string | null
          prospect_num?: string | null
          stage?: string | null
          update_today?: string | null
          updated_at?: string
          waiting_for_reply?: boolean | null
        }
        Relationships: []
      }
      ai_whatsapp_session_nodepath: {
        Row: {
          id_device: string
          id_prospect: string
          id_sessionx: number
          timestamp: string
        }
        Insert: {
          id_device: string
          id_prospect: string
          id_sessionx?: number
          timestamp: string
        }
        Update: {
          id_device?: string
          id_prospect?: string
          id_sessionx?: number
          timestamp?: string
        }
        Relationships: []
      }
      chatbot_flows_nodepath: {
        Row: {
          created_at: string
          edges: Json | null
          id: string
          id_device: string
          name: string
          niche: string
          nodes: Json | null
          updated_at: string
        }
        Insert: {
          created_at?: string
          edges?: Json | null
          id: string
          id_device?: string
          name: string
          niche?: string
          nodes?: Json | null
          updated_at?: string
        }
        Update: {
          created_at?: string
          edges?: Json | null
          id?: string
          id_device?: string
          name?: string
          niche?: string
          nodes?: Json | null
          updated_at?: string
        }
        Relationships: []
      }
      device_setting_nodepath: {
        Row: {
          api_key: string | null
          api_key_option:
            | Database["public"]["Enums"]["api_key_option_enum"]
            | null
          created_at: string
          device_id: string | null
          id: string
          id_admin: string | null
          id_device: string | null
          id_erp: string | null
          instance: string | null
          phone_number: string | null
          provider: Database["public"]["Enums"]["provider_enum"] | null
          updated_at: string
          user_id: string | null
          webhook_id: string | null
        }
        Insert: {
          api_key?: string | null
          api_key_option?:
            | Database["public"]["Enums"]["api_key_option_enum"]
            | null
          created_at?: string
          device_id?: string | null
          id: string
          id_admin?: string | null
          id_device?: string | null
          id_erp?: string | null
          instance?: string | null
          phone_number?: string | null
          provider?: Database["public"]["Enums"]["provider_enum"] | null
          updated_at?: string
          user_id?: string | null
          webhook_id?: string | null
        }
        Update: {
          api_key?: string | null
          api_key_option?:
            | Database["public"]["Enums"]["api_key_option_enum"]
            | null
          created_at?: string
          device_id?: string | null
          id?: string
          id_admin?: string | null
          id_device?: string | null
          id_erp?: string | null
          instance?: string | null
          phone_number?: string | null
          provider?: Database["public"]["Enums"]["provider_enum"] | null
          updated_at?: string
          user_id?: string | null
          webhook_id?: string | null
        }
        Relationships: []
      }
      orders_nodepath: {
        Row: {
          amount: number
          bill_id: string | null
          collection_id: string | null
          created_at: string
          id: number
          method: string | null
          product: string
          status: string | null
          updated_at: string
          url: string | null
          user_id: string | null
        }
        Insert: {
          amount: number
          bill_id?: string | null
          collection_id?: string | null
          created_at?: string
          id?: number
          method?: string | null
          product: string
          status?: string | null
          updated_at?: string
          url?: string | null
          user_id?: string | null
        }
        Update: {
          amount?: number
          bill_id?: string | null
          collection_id?: string | null
          created_at?: string
          id?: number
          method?: string | null
          product?: string
          status?: string | null
          updated_at?: string
          url?: string | null
          user_id?: string | null
        }
        Relationships: []
      }
      profiles: {
        Row: {
          created_at: string
          expired: string | null
          full_name: string
          gmail: string | null
          id: string
          last_login: string | null
          phone: string | null
          status: string | null
          updated_at: string
        }
        Insert: {
          created_at?: string
          expired?: string | null
          full_name: string
          gmail?: string | null
          id: string
          last_login?: string | null
          phone?: string | null
          status?: string | null
          updated_at?: string
        }
        Update: {
          created_at?: string
          expired?: string | null
          full_name?: string
          gmail?: string | null
          id?: string
          last_login?: string | null
          phone?: string | null
          status?: string | null
          updated_at?: string
        }
        Relationships: []
      }
      stagesetvalue_nodepath: {
        Row: {
          columnsdata: string | null
          id_device: string | null
          inputhardcode: string | null
          stage: string | null
          stagesetvalue_id: number
          type_inputdata: string | null
        }
        Insert: {
          columnsdata?: string | null
          id_device?: string | null
          inputhardcode?: string | null
          stage?: string | null
          stagesetvalue_id?: number
          type_inputdata?: string | null
        }
        Update: {
          columnsdata?: string | null
          id_device?: string | null
          inputhardcode?: string | null
          stage?: string | null
          stagesetvalue_id?: number
          type_inputdata?: string | null
        }
        Relationships: []
      }
      user_nodepath: {
        Row: {
          created_at: string
          email: string
          expired: string | null
          full_name: string
          gmail: string | null
          id: string
          is_active: boolean | null
          last_login: string | null
          password: string
          phone: string | null
          status: string | null
          updated_at: string
        }
        Insert: {
          created_at?: string
          email: string
          expired?: string | null
          full_name: string
          gmail?: string | null
          id: string
          is_active?: boolean | null
          last_login?: string | null
          password: string
          phone?: string | null
          status?: string | null
          updated_at?: string
        }
        Update: {
          created_at?: string
          email?: string
          expired?: string | null
          full_name?: string
          gmail?: string | null
          id?: string
          is_active?: boolean | null
          last_login?: string | null
          password?: string
          phone?: string | null
          status?: string | null
          updated_at?: string
        }
        Relationships: []
      }
      user_sessions_nodepath: {
        Row: {
          created_at: string
          expires_at: string
          id: string
          token: string
          user_id: string
        }
        Insert: {
          created_at?: string
          expires_at?: string
          id: string
          token: string
          user_id: string
        }
        Update: {
          created_at?: string
          expires_at?: string
          id?: string
          token?: string
          user_id?: string
        }
        Relationships: []
      }
      wasapbot_nodepath: {
        Row: {
          alamat: string | null
          cara_bayaran: string | null
          conv_last: string | null
          conv_start: string | null
          current_node_id: string | null
          date_last: string | null
          date_start: string | null
          execution_id: string | null
          execution_status: string | null
          flow_id: string | null
          flow_reference: string | null
          id_device: string | null
          id_prospect: number
          last_node_id: string | null
          nama: string | null
          niche: string | null
          no_fon: string | null
          pakej: string | null
          peringkat_sekolah: string | null
          prospect_num: string | null
          stage: string | null
          status: string | null
          tarikh_gaji: string | null
          temp_stage: string | null
          waiting_for_reply: boolean | null
        }
        Insert: {
          alamat?: string | null
          cara_bayaran?: string | null
          conv_last?: string | null
          conv_start?: string | null
          current_node_id?: string | null
          date_last?: string | null
          date_start?: string | null
          execution_id?: string | null
          execution_status?: string | null
          flow_id?: string | null
          flow_reference?: string | null
          id_device?: string | null
          id_prospect?: number
          last_node_id?: string | null
          nama?: string | null
          niche?: string | null
          no_fon?: string | null
          pakej?: string | null
          peringkat_sekolah?: string | null
          prospect_num?: string | null
          stage?: string | null
          status?: string | null
          tarikh_gaji?: string | null
          temp_stage?: string | null
          waiting_for_reply?: boolean | null
        }
        Update: {
          alamat?: string | null
          cara_bayaran?: string | null
          conv_last?: string | null
          conv_start?: string | null
          current_node_id?: string | null
          date_last?: string | null
          date_start?: string | null
          execution_id?: string | null
          execution_status?: string | null
          flow_id?: string | null
          flow_reference?: string | null
          id_device?: string | null
          id_prospect?: number
          last_node_id?: string | null
          nama?: string | null
          niche?: string | null
          no_fon?: string | null
          pakej?: string | null
          peringkat_sekolah?: string | null
          prospect_num?: string | null
          stage?: string | null
          status?: string | null
          tarikh_gaji?: string | null
          temp_stage?: string | null
          waiting_for_reply?: boolean | null
        }
        Relationships: []
      }
      wasapbot_session_nodepath: {
        Row: {
          id_device: string
          id_prospect: string
          id_sessiony: number
          session_timestamp: string
        }
        Insert: {
          id_device: string
          id_prospect: string
          id_sessiony?: number
          session_timestamp: string
        }
        Update: {
          id_device?: string
          id_prospect?: string
          id_sessiony?: number
          session_timestamp?: string
        }
        Relationships: []
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      [_ in never]: never
    }
    Enums: {
      api_key_option_enum:
        | "openai/gpt-5-chat"
        | "openai/gpt-5-mini"
        | "openai/chatgpt-4o-latest"
        | "openai/gpt-4.1"
        | "google/gemini-2.5-pro"
        | "google/gemini-pro-1.5"
      provider_enum: "whacenter" | "wablas" | "waha"
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

type DatabaseWithoutInternals = Omit<Database, "__InternalSupabase">

type DefaultSchema = DatabaseWithoutInternals[Extract<keyof Database, "public">]

export type Tables<
  DefaultSchemaTableNameOrOptions extends
    | keyof (DefaultSchema["Tables"] & DefaultSchema["Views"])
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
        DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
      DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])[TableName] extends {
      Row: infer R
    }
    ? R
    : never
  : DefaultSchemaTableNameOrOptions extends keyof (DefaultSchema["Tables"] &
        DefaultSchema["Views"])
    ? (DefaultSchema["Tables"] &
        DefaultSchema["Views"])[DefaultSchemaTableNameOrOptions] extends {
        Row: infer R
      }
      ? R
      : never
    : never

export type TablesInsert<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Insert: infer I
    }
    ? I
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Insert: infer I
      }
      ? I
      : never
    : never

export type TablesUpdate<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Update: infer U
    }
    ? U
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Update: infer U
      }
      ? U
      : never
    : never

export type Enums<
  DefaultSchemaEnumNameOrOptions extends
    | keyof DefaultSchema["Enums"]
    | { schema: keyof DatabaseWithoutInternals },
  EnumName extends DefaultSchemaEnumNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"]
    : never = never,
> = DefaultSchemaEnumNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"][EnumName]
  : DefaultSchemaEnumNameOrOptions extends keyof DefaultSchema["Enums"]
    ? DefaultSchema["Enums"][DefaultSchemaEnumNameOrOptions]
    : never

export type CompositeTypes<
  PublicCompositeTypeNameOrOptions extends
    | keyof DefaultSchema["CompositeTypes"]
    | { schema: keyof DatabaseWithoutInternals },
  CompositeTypeName extends PublicCompositeTypeNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"]
    : never = never,
> = PublicCompositeTypeNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"][CompositeTypeName]
  : PublicCompositeTypeNameOrOptions extends keyof DefaultSchema["CompositeTypes"]
    ? DefaultSchema["CompositeTypes"][PublicCompositeTypeNameOrOptions]
    : never

export const Constants = {
  public: {
    Enums: {
      api_key_option_enum: [
        "openai/gpt-5-chat",
        "openai/gpt-5-mini",
        "openai/chatgpt-4o-latest",
        "openai/gpt-4.1",
        "google/gemini-2.5-pro",
        "google/gemini-pro-1.5",
      ],
      provider_enum: ["whacenter", "wablas", "waha"],
    },
  },
} as const
