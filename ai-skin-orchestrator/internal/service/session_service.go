package service

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ConversationMessage represents a message in conversation history
type ConversationMessage struct {
	Role      string    `json:"role"` // "user" or "assistant" or "bot"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a conversation session
type Session struct {
	SessionID    string
	UserID       string
	Channel      string
	Messages     []ConversationMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
	mu           sync.RWMutex
}

// SessionService manages conversation sessions
type SessionService struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewSessionService creates a new session service
func NewSessionService() *SessionService {
	return &SessionService{
		sessions: make(map[string]*Session),
		ttl:      24 * time.Hour, // 24 hour TTL
	}
}

// GetOrCreateSession gets an existing session or creates a new one
func (ss *SessionService) GetOrCreateSession(ctx context.Context, sessionID, userID, channel string) *Session {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if sessionID == "" {
		// Create new session
		sessionID = generateSessionID()
	}

	session, exists := ss.sessions[sessionID]
	if !exists {
		session = &Session{
			SessionID: sessionID,
			UserID:    userID,
			Channel:   channel,
			Messages:  []ConversationMessage{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		ss.sessions[sessionID] = session
		log.Info().Str("session_id", sessionID).Str("user_id", userID).Msg("Created new session")
	}

	// Check if session expired
	if time.Since(session.UpdatedAt) > ss.ttl {
		delete(ss.sessions, sessionID)
		session = &Session{
			SessionID: sessionID,
			UserID:    userID,
			Channel:   channel,
			Messages:  []ConversationMessage{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		ss.sessions[sessionID] = session
		log.Info().Str("session_id", sessionID).Msg("Session expired, created new session")
	}

	return session
}

// AddMessage adds a message to the session
func (ss *SessionService) AddMessage(sessionID, role, content string) {
	ss.mu.RLock()
	session, exists := ss.sessions[sessionID]
	ss.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.Messages = append(session.Messages, ConversationMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	session.UpdatedAt = time.Now()

	// Keep only last 20 messages to prevent memory bloat
	if len(session.Messages) > 20 {
		session.Messages = session.Messages[len(session.Messages)-20:]
	}
}

// GetConversationHistory returns conversation history in format for LLM
func (ss *SessionService) GetConversationHistory(sessionID string) []map[string]string {
	ss.mu.RLock()
	session, exists := ss.sessions[sessionID]
	ss.mu.RUnlock()

	if !exists {
		return []map[string]string{}
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	history := make([]map[string]string, 0, len(session.Messages))
	for _, msg := range session.Messages {
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	return history
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return "session_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

