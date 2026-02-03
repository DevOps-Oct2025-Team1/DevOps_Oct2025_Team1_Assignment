import type { User, AIModel, Conversation } from "./types"

export const aiModels: AIModel[] = [
  {
    id: "gemma3",
    name: "Gemma 3",
    provider: "Google",
    description: "Lightweight, state-of-the-art open model from Google",
    icon: "🔵",
    maxTokens: 8192,
  },
  {
    id: "qwen3",
    name: "Qwen 3",
    provider: "Alibaba",
    description: "Advanced multilingual model with strong reasoning",
    icon: "🟣",
    maxTokens: 8192,
  },
]

// Helper to get model info by id
export const getModelById = (id: string): AIModel | undefined => {
  return aiModels.find(model => model.id === id)
}

// Helper to create AIModel from model id string
export const createModelFromId = (id: string): AIModel => {
  const existing = getModelById(id)
  if (existing) return existing

  // Fallback for unknown models
  return {
    id,
    name: id.charAt(0).toUpperCase() + id.slice(1),
    provider: "Unknown",
    description: "AI model",
    icon: "🤖",
    maxTokens: 4096
  }
}

export const initialUsers: User[] = [
  {
    id: "1",
    name: "Alex Johnson",
    email: "alex@company.com",
    role: "admin",
    status: "active",
    createdAt: "2024-01-15",
    lastActive: "2024-01-20",
  },
  {
    id: "2",
    name: "Sarah Chen",
    email: "sarah@company.com",
    role: "user",
    status: "active",
    createdAt: "2024-01-10",
    lastActive: "2024-01-19",
  },
  {
    id: "3",
    name: "Mike Williams",
    email: "mike@company.com",
    role: "user",
    status: "inactive",
    createdAt: "2024-01-05",
    lastActive: "2024-01-12",
  },
  {
    id: "4",
    name: "Emily Davis",
    email: "emily@company.com",
    role: "user",
    status: "active",
    createdAt: "2024-01-08",
    lastActive: "2024-01-20",
  },
  {
    id: "5",
    name: "James Brown",
    email: "james@company.com",
    role: "user",
    status: "suspended",
    createdAt: "2023-12-20",
    lastActive: "2024-01-02",
  },
]

export const sampleConversations: Conversation[] = [
  {
    id: "1",
    title: "Code Review Help",
    model: "gpt-4",
    messages: [],
    createdAt: new Date("2024-01-19"),
    updatedAt: new Date("2024-01-19"),
  },
  {
    id: "2",
    title: "Marketing Strategy",
    model: "claude-3",
    messages: [],
    createdAt: new Date("2024-01-18"),
    updatedAt: new Date("2024-01-18"),
  },
  {
    id: "3",
    title: "Data Analysis",
    model: "gemini-pro",
    messages: [],
    createdAt: new Date("2024-01-17"),
    updatedAt: new Date("2024-01-17"),
  },
]
