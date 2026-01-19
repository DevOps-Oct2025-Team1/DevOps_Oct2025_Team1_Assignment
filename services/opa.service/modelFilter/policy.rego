package ai.access

default allowed_models := []

role_to_models := {
  "admin": ["deekseek-r1", "qwen-3.2", "llama-3.2"],
  "premium": ["qwen-3.2", "llama-3.2"],
  "free": ["llama-3.2"]
}

allowed_models := models if {
  role := input.user.role
  models := role_to_models[role]
}

