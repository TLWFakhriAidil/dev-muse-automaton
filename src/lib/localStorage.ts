import { ChatbotFlow, MediaFile, FlowExecution } from '@/types/chatbot'
import { saveFlow as saveSupabaseFlow, getFlows as getSupabaseFlows, getFlow as getSupabaseFlow, deleteFlow as deleteSupabaseFlow } from './supabaseFlowStorage'

const MEDIA_KEY = 'chatbot_media'

// Flow management - using Supabase
export const saveFlow = async (flow: ChatbotFlow): Promise<void> => {
  return saveSupabaseFlow(flow)
}

export const getFlows = async (): Promise<ChatbotFlow[]> => {
  return getSupabaseFlows()
}

export const getFlow = async (id: string): Promise<ChatbotFlow | null> => {
  return getSupabaseFlow(id)
}

export const deleteFlow = async (id: string): Promise<void> => {
  return deleteSupabaseFlow(id)
}

// Media management
export const saveMediaFile = (file: MediaFile): void => {
  const media = getMediaFiles()
  media.push(file)
  localStorage.setItem(MEDIA_KEY, JSON.stringify(media))
}

export const getMediaFiles = (): MediaFile[] => {
  const stored = localStorage.getItem(MEDIA_KEY)
  return stored ? JSON.parse(stored) : []
}

export const getMediaFile = (id: string): MediaFile | null => {
  const media = getMediaFiles()
  return media.find(m => m.id === id) || null
}

export const deleteMediaFile = (id: string): void => {
  const media = getMediaFiles().filter(m => m.id !== id)
  localStorage.setItem(MEDIA_KEY, JSON.stringify(media))
}

// Execution management - using local storage for now
const EXECUTION_KEY = 'chatbot_executions'

export const saveFlowExecution = async (execution: FlowExecution): Promise<void> => {
  const executions = getFlowExecutionsFromStorage()
  const existingIndex = executions.findIndex(e => e.id === execution.id)
  
  if (existingIndex >= 0) {
    executions[existingIndex] = execution
  } else {
    executions.push(execution)
  }
  
  localStorage.setItem(EXECUTION_KEY, JSON.stringify(executions))
}

export const getFlowExecution = async (flowId: string): Promise<FlowExecution | null> => {
  const executions = getFlowExecutionsFromStorage()
  return executions.find(e => e.flowId === flowId) || null
}

export const updateFlowExecution = async (execution: FlowExecution): Promise<void> => {
  return saveFlowExecution(execution)
}

export const deleteFlowExecution = async (flowId: string): Promise<void> => {
  const executions = getFlowExecutionsFromStorage()
  const filtered = executions.filter(e => e.flowId !== flowId)
  localStorage.setItem(EXECUTION_KEY, JSON.stringify(filtered))
}

const getFlowExecutionsFromStorage = (): FlowExecution[] => {
  try {
    const stored = localStorage.getItem(EXECUTION_KEY)
    return stored ? JSON.parse(stored) : []
  } catch (error) {
    console.error('Error parsing flow executions from localStorage:', error)
    return []
  }
}

// Utility functions
export const createMediaFileFromFile = async (file: File): Promise<MediaFile> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      
      const mediaFile: MediaFile = {
        id: `media_${Date.now()}_${Math.random().toString(36).substring(2)}`,
        filename: file.name,
        type: file.type,
        size: file.size,
        dataUrl,
        createdAt: new Date().toISOString()
      }
      
      resolve(mediaFile)
    }
    
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsDataURL(file)
  })
}

export const replaceVariables = (text: string, variables: Record<string, string>): string => {
  let result = text
  Object.entries(variables).forEach(([key, value]) => {
    result = result.replace(new RegExp(`{{${key}}}`, 'g'), value)
  })
  return result
}