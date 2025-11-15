package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateTaskID generates a task ID with prefix
func GenerateTaskID() string {
	return "task_" + uuid.New().String()
}

// GenerateSessionID generates a session ID with prefix
func GenerateSessionID() string {
	return "sess_" + uuid.New().String()
}

// GenerateAgentID generates an agent ID with prefix
func GenerateAgentID() string {
	return "agent_" + uuid.New().String()
}

