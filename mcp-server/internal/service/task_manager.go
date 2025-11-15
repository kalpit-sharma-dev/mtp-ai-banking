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

// TaskManager handles task lifecycle management
type TaskManager struct {
	redisClient    *redis.Client
	redisAvailable bool
	tasks          map[string]*model.Task // In-memory fallback
	mu             sync.RWMutex
	ttl            time.Duration
}

// NewTaskManager creates a new task manager instance
func NewTaskManager(redisClient *redis.Client) *TaskManager {
	tm := &TaskManager{
		redisClient: redisClient,
		tasks:       make(map[string]*model.Task),
		ttl:         7 * 24 * time.Hour, // 7 days TTL for tasks
	}

	// Check Redis availability
	ctx := context.Background()
	if redisClient != nil {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			tm.redisAvailable = true
		} else {
			tm.redisAvailable = false
			log.Warn().Msg("Redis unavailable for tasks, using in-memory storage only")
		}
	} else {
		tm.redisAvailable = false
		log.Warn().Msg("Redis client not provided for tasks, using in-memory storage only")
	}

	return tm
}

// CreateTask creates a new task from a request
func (tm *TaskManager) CreateTask(ctx context.Context, req *model.TaskRequest, sessionID string) (*model.Task, error) {
	taskID := utils.GenerateTaskID()

	task := &model.Task{
		TaskID:    taskID,
		SessionID: sessionID,
		UserID:    req.UserID,
		Channel:   req.Channel,
		Intent:    req.Intent,
		Status:    model.TaskStatusPending,
		Data:      req.Data,
		Context:   req.Context,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if task.Context == nil {
		task.Context = make(map[string]interface{})
	}

	// Save to Redis (if available)
	if tm.redisAvailable {
		if err := tm.saveTask(ctx, task); err != nil {
			log.Warn().Err(err).Msg("Failed to save task to Redis, using in-memory storage")
			tm.redisAvailable = false
		}
	}

	// Always store in memory
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	log.Info().
		Str("task_id", taskID).
		Str("session_id", sessionID).
		Str("user_id", req.UserID).
		Str("intent", req.Intent).
		Msg("Task created")

	return task, nil
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	// Try in-memory first
	tm.mu.RLock()
	if task, ok := tm.tasks[taskID]; ok {
		tm.mu.RUnlock()
		return task, nil
	}
	tm.mu.RUnlock()

	// Fallback to Redis (if available)
	if tm.redisAvailable && tm.redisClient != nil {
		key := fmt.Sprintf("task:%s", taskID)
		data, err := tm.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		if err != nil {
			tm.redisAvailable = false
			return nil, fmt.Errorf("task not found: %s", taskID)
		}

		var task model.Task
		if err := json.Unmarshal([]byte(data), &task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal task: %w", err)
		}

		// Cache in memory
		tm.mu.Lock()
		tm.tasks[taskID] = &task
		tm.mu.Unlock()

		return &task, nil
	}

	return nil, fmt.Errorf("task not found: %s", taskID)
}

// UpdateTaskStatus updates the task status and optionally result
func (tm *TaskManager) UpdateTaskStatus(ctx context.Context, taskID string, status model.TaskStatus, result map[string]interface{}, errorMsg string) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if result != nil {
		task.Result = result
	}

	if errorMsg != "" {
		task.Error = errorMsg
	}

	if status == model.TaskStatusCompleted || status == model.TaskStatusFailed || status == model.TaskStatusRejected {
		now := time.Now()
		task.CompletedAt = &now
	}

	// Save to Redis (if available)
	if tm.redisAvailable {
		if err := tm.saveTask(ctx, task); err != nil {
			log.Warn().Err(err).Msg("Failed to save task update to Redis")
			tm.redisAvailable = false
		}
	}

	// Always update in memory
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	return nil
}

// UpdateTaskAgent updates the agent assigned to a task
func (tm *TaskManager) UpdateTaskAgent(ctx context.Context, taskID, agentID string) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.AgentID = agentID
	task.Status = model.TaskStatusProcessing
	task.UpdatedAt = time.Now()

	// Save to Redis (if available)
	if tm.redisAvailable {
		if err := tm.saveTask(ctx, task); err != nil {
			log.Warn().Err(err).Msg("Failed to save task agent update to Redis")
			tm.redisAvailable = false
		}
	}

	// Always update in memory
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	return nil
}

// UpdateTaskResult updates task with final result, risk score, and explanation
func (tm *TaskManager) UpdateTaskResult(ctx context.Context, taskID string, result map[string]interface{}, riskScore float64, explanation string) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.Result = result
	task.RiskScore = riskScore
	task.Explanation = explanation
	task.Status = model.TaskStatusCompleted
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = now

	// Save to Redis (if available)
	if tm.redisAvailable {
		if err := tm.saveTask(ctx, task); err != nil {
			log.Warn().Err(err).Msg("Failed to save task result to Redis")
			tm.redisAvailable = false
		}
	}

	// Always update in memory
	tm.mu.Lock()
	tm.tasks[taskID] = task
	tm.mu.Unlock()

	return nil
}

// saveTask saves task to Redis
func (tm *TaskManager) saveTask(ctx context.Context, task *model.Task) error {
	if tm.redisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("task:%s", task.TaskID)
	
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := tm.redisClient.Set(ctx, key, data, tm.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set task in Redis: %w", err)
	}

	return nil
}

