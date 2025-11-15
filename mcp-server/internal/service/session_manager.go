package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aibanking/mcp-server/internal/model"
	"github.com/aibanking/mcp-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// SessionManager handles session creation, retrieval, and context management
type SessionManager struct {
	redisClient    *redis.Client
	redisAvailable bool
	sessions       map[string]*model.Session // In-memory fallback
	mu             sync.RWMutex
	ttl            time.Duration
}

// NewSessionManager creates a new session manager instance
func NewSessionManager(redisClient *redis.Client) *SessionManager {
	sm := &SessionManager{
		redisClient: redisClient,
		sessions:    make(map[string]*model.Session),
		ttl:         24 * time.Hour, // Default 24 hour TTL
	}

	// Check Redis availability
	ctx := context.Background()
	if redisClient != nil {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			sm.redisAvailable = true
		} else {
			sm.redisAvailable = false
			log.Warn().Msg("Redis unavailable for sessions, using in-memory storage only")
		}
	} else {
		sm.redisAvailable = false
		log.Warn().Msg("Redis client not provided for sessions, using in-memory storage only")
	}

	return sm
}

// CreateSession creates a new session with context
func (sm *SessionManager) CreateSession(ctx context.Context, req *model.SessionRequest) (*model.Session, error) {
	sessionID := utils.GenerateSessionID()
	
	session := &model.Session{
		SessionID:   sessionID,
		UserID:      req.UserID,
		Channel:     req.Channel,
		Context:     req.Context,
		Metadata:    make(map[string]interface{}),
		TaskHistory: []string{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(sm.ttl),
	}

	if session.Context == nil {
		session.Context = make(map[string]interface{})
	}

	// Store session in Redis (if available)
	if sm.redisAvailable {
		if err := sm.saveSession(ctx, session); err != nil {
			log.Warn().Err(err).Msg("Failed to save session to Redis, using in-memory storage")
			sm.redisAvailable = false
		}
	}

	// Always store in memory
	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	log.Info().
		Str("session_id", sessionID).
		Str("user_id", req.UserID).
		Str("channel", req.Channel).
		Msg("Session created")

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	// Try in-memory first
	sm.mu.RLock()
	if session, ok := sm.sessions[sessionID]; ok {
		sm.mu.RUnlock()
		// Check if expired
		if time.Now().After(session.ExpiresAt) {
			sm.mu.Lock()
			delete(sm.sessions, sessionID)
			sm.mu.Unlock()
			return nil, fmt.Errorf("session expired: %s", sessionID)
		}
		return session, nil
	}
	sm.mu.RUnlock()

	// Fallback to Redis (if available)
	if sm.redisAvailable && sm.redisClient != nil {
		key := fmt.Sprintf("session:%s", sessionID)
		data, err := sm.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		if err != nil {
			sm.redisAvailable = false
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}

		var session model.Session
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			return nil, fmt.Errorf("failed to unmarshal session: %w", err)
		}

		// Check if session expired
		if time.Now().After(session.ExpiresAt) {
			sm.redisClient.Del(ctx, key)
			return nil, fmt.Errorf("session expired: %s", sessionID)
		}

		// Cache in memory
		sm.mu.Lock()
		sm.sessions[sessionID] = &session
		sm.mu.Unlock()

		return &session, nil
	}

	return nil, fmt.Errorf("session not found: %s", sessionID)
}

// UpdateSession updates session context and metadata
func (sm *SessionManager) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Update context if provided
	if ctxData, ok := updates["context"].(map[string]interface{}); ok {
		for k, v := range ctxData {
			session.Context[k] = v
		}
	}

	// Update metadata if provided
	if metaData, ok := updates["metadata"].(map[string]interface{}); ok {
		for k, v := range metaData {
			session.Metadata[k] = v
		}
	}

	// Add task to history if provided
	if taskID, ok := updates["task_id"].(string); ok {
		session.TaskHistory = append(session.TaskHistory, taskID)
	}

	session.UpdatedAt = time.Now()

	// Save to Redis (if available)
	if sm.redisAvailable {
		if err := sm.saveSession(ctx, session); err != nil {
			log.Warn().Err(err).Msg("Failed to save session update to Redis")
			sm.redisAvailable = false
		}
	}

	// Always update in memory
	sm.mu.Lock()
	sm.sessions[session.SessionID] = session
	sm.mu.Unlock()

	return nil
}

// AddTaskToSession adds a task ID to the session's task history
func (sm *SessionManager) AddTaskToSession(ctx context.Context, sessionID, taskID string) error {
	return sm.UpdateSession(ctx, sessionID, map[string]interface{}{
		"task_id": taskID,
	})
}

// saveSession saves session to Redis
func (sm *SessionManager) saveSession(ctx context.Context, session *model.Session) error {
	if sm.redisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("session:%s", session.SessionID)
	
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if err := sm.redisClient.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set session in Redis: %w", err)
	}

	return nil
}

