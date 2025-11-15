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

// AgentRegistry manages agent registration and discovery
type AgentRegistry struct {
	redisClient   *redis.Client
	redisAvailable bool
	mu            sync.RWMutex
	agents        map[string]*model.Agent // In-memory cache
}

// NewAgentRegistry creates a new agent registry instance
func NewAgentRegistry(redisClient *redis.Client) *AgentRegistry {
	registry := &AgentRegistry{
		redisClient: redisClient,
		agents:      make(map[string]*model.Agent),
	}
	
	// Check Redis availability
	ctx := context.Background()
	if redisClient != nil {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			registry.redisAvailable = true
			// Load agents from Redis on startup
			go registry.loadAgentsFromRedis(ctx)
		} else {
			registry.redisAvailable = false
			log.Warn().Msg("Redis unavailable, using in-memory storage only")
		}
	} else {
		registry.redisAvailable = false
		log.Warn().Msg("Redis client not provided, using in-memory storage only")
	}
	
	return registry
}

// RegisterAgent registers a new agent in the mesh
func (ar *AgentRegistry) RegisterAgent(ctx context.Context, req *model.AgentRegistrationRequest) (*model.Agent, error) {
	agentID := utils.GenerateAgentID()
	now := time.Now()

	agent := &model.Agent{
		AgentID:      agentID,
		Name:         req.Name,
		Type:         model.AgentType(req.Type),
		Endpoint:     req.Endpoint,
		GRPCEndpoint: req.GRPCEndpoint,
		Status:       model.AgentStatusHealthy,
		Capabilities: req.Capabilities,
		Rules:        req.Rules,
		Metadata:     req.Metadata,
		HealthCheck:  req.HealthCheck,
		LastHealthAt: now,
		RegisteredAt: now,
		UpdatedAt:    now,
	}

	if agent.Rules == nil {
		agent.Rules = make(map[string]interface{})
	}
	if agent.Metadata == nil {
		agent.Metadata = make(map[string]interface{})
	}

	// Save to Redis (if available)
	if ar.redisAvailable {
		if err := ar.saveAgent(ctx, agent); err != nil {
			log.Warn().Err(err).Msg("Failed to save agent to Redis, continuing with in-memory storage")
			ar.redisAvailable = false // Mark Redis as unavailable
		}
	}

	// Update in-memory cache (always)
	ar.mu.Lock()
	ar.agents[agentID] = agent
	ar.mu.Unlock()

	log.Info().
		Str("agent_id", agentID).
		Str("name", req.Name).
		Str("type", req.Type).
		Msg("Agent registered")

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (ar *AgentRegistry) GetAgent(ctx context.Context, agentID string) (*model.Agent, error) {
	// Try in-memory cache first
	ar.mu.RLock()
	if agent, ok := ar.agents[agentID]; ok {
		ar.mu.RUnlock()
		return agent, nil
	}
	ar.mu.RUnlock()

	// Fallback to Redis (if available)
	if ar.redisAvailable && ar.redisClient != nil {
		key := fmt.Sprintf("agent:%s", agentID)
		data, err := ar.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
		if err != nil {
			// Redis error, mark as unavailable and return not found
			ar.redisAvailable = false
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}

		var agent model.Agent
		if err := json.Unmarshal([]byte(data), &agent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal agent: %w", err)
		}

		// Update cache
		ar.mu.Lock()
		ar.agents[agentID] = &agent
		ar.mu.Unlock()

		return &agent, nil
	}

	return nil, fmt.Errorf("agent not found: %s", agentID)
}

// FindAgentsByType finds all agents of a specific type
func (ar *AgentRegistry) FindAgentsByType(ctx context.Context, agentType model.AgentType) ([]*model.Agent, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var agents []*model.Agent
	for _, agent := range ar.agents {
		if agent.Type == agentType && agent.Status == model.AgentStatusHealthy {
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// FindAgentsByCapability finds agents that can handle a specific capability
func (ar *AgentRegistry) FindAgentsByCapability(ctx context.Context, capability string) ([]*model.Agent, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var agents []*model.Agent
	for _, agent := range ar.agents {
		if agent.Status != model.AgentStatusHealthy {
			continue
		}
		for _, cap := range agent.Capabilities {
			if cap == capability {
				agents = append(agents, agent)
				break
			}
		}
	}

	return agents, nil
}

// UpdateAgentStatus updates the health status of an agent
func (ar *AgentRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status model.AgentStatus) error {
	agent, err := ar.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	agent.Status = status
	agent.LastHealthAt = time.Now()
	agent.UpdatedAt = time.Now()

	// Save to Redis (if available)
	if ar.redisAvailable {
		if err := ar.saveAgent(ctx, agent); err != nil {
			log.Warn().Err(err).Msg("Failed to save agent status to Redis")
			ar.redisAvailable = false
		}
	}

	// Update cache (always)
	ar.mu.Lock()
	ar.agents[agentID] = agent
	ar.mu.Unlock()

	return nil
}

// GetAllAgents returns all registered agents
func (ar *AgentRegistry) GetAllAgents(ctx context.Context) ([]*model.Agent, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	agents := make([]*model.Agent, 0, len(ar.agents))
	for _, agent := range ar.agents {
		agents = append(agents, agent)
	}

	return agents, nil
}

// saveAgent saves agent to Redis
func (ar *AgentRegistry) saveAgent(ctx context.Context, agent *model.Agent) error {
	if ar.redisClient == nil {
		return fmt.Errorf("redis client not available")
	}

	key := fmt.Sprintf("agent:%s", agent.AgentID)
	
	data, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	if err := ar.redisClient.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to set agent in Redis: %w", err)
	}

	// Also add to set of all agent IDs
	if err := ar.redisClient.SAdd(ctx, "agents:all", agent.AgentID).Err(); err != nil {
		return fmt.Errorf("failed to add agent to set: %w", err)
	}

	return nil
}

// loadAgentsFromRedis loads all agents from Redis into memory
func (ar *AgentRegistry) loadAgentsFromRedis(ctx context.Context) {
	if ar.redisClient == nil || !ar.redisAvailable {
		return
	}

	agentIDs, err := ar.redisClient.SMembers(ctx, "agents:all").Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load agents from Redis")
		ar.redisAvailable = false
		return
	}

	for _, agentID := range agentIDs {
		key := fmt.Sprintf("agent:%s", agentID)
		data, err := ar.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var agent model.Agent
		if err := json.Unmarshal([]byte(data), &agent); err != nil {
			continue
		}

		ar.mu.Lock()
		ar.agents[agentID] = &agent
		ar.mu.Unlock()
	}

	if len(agentIDs) > 0 {
		log.Info().Int("count", len(agentIDs)).Msg("Loaded agents from Redis")
	}
}

