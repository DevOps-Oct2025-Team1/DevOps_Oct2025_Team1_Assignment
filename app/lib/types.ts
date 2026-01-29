export interface User {
  id: string
  name: string
  email: string
  role: "user" | "admin"
  status: "active" | "inactive" | "suspended"
  createdAt: string
  lastActive: string
  avatar?: string
}

export interface Message {
  id: string
  role: "user" | "assistant"
  content: string
  timestamp: Date
  model?: string
}

export interface AIModel {
  id: string
  name: string
  provider: string
  description: string
  icon: string
  maxTokens: number
}

export interface Conversation {
  id: string
  title: string
  messages: Message[]
  model: string
  createdAt: Date
  updatedAt: Date
}
