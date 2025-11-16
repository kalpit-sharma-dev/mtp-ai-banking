package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/aibanking/ai-skin-orchestrator/internal/model"
	"github.com/rs/zerolog/log"
)

// Document represents a document in the vector store
type Document struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Embedding   []float64              `json:"embedding"`
	Metadata    map[string]interface{} `json:"metadata"`
	Type        string                 `json:"type"` // "conversation", "transaction", "account", "policy"
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Score       float64                `json:"score,omitempty"` // For retrieval results
}

// RAGService provides Retrieval-Augmented Generation capabilities
type RAGService struct {
	documents map[string]*Document // In-memory storage (can be replaced with vector DB)
	mu        sync.RWMutex
	llmService *LLMService
	embeddingCache map[string][]float64 // Cache for embeddings
	embeddingMu    sync.RWMutex
}

// NewRAGService creates a new RAG service
func NewRAGService(llmService *LLMService) *RAGService {
	return &RAGService{
		documents:      make(map[string]*Document),
		llmService:     llmService,
		embeddingCache: make(map[string][]float64),
	}
}

// StoreConversation stores a conversation message in the vector store
func (r *RAGService) StoreConversation(ctx context.Context, userID, sessionID, role, content string) error {
	docID := fmt.Sprintf("conv_%s_%d", sessionID, time.Now().UnixNano())
	
	// Create document
	doc := &Document{
		ID:        docID,
		Content:   content,
		Metadata: map[string]interface{}{
			"role":      role,
			"user_id":   userID,
			"session_id": sessionID,
		},
		Type:      "conversation",
		UserID:    userID,
		SessionID: sessionID,
		Timestamp: time.Now(),
	}

	// Generate embedding
	embedding, err := r.generateEmbedding(ctx, content)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate embedding, using TF-IDF fallback")
		embedding = r.generateTFIDFEmbedding(content)
	}
	doc.Embedding = embedding

	// Store document
	r.mu.Lock()
	r.documents[docID] = doc
	r.mu.Unlock()

	log.Debug().
		Str("doc_id", docID).
		Str("user_id", userID).
		Str("type", "conversation").
		Msg("Stored conversation in RAG")

	return nil
}

// StoreTransaction stores a transaction record in the vector store
func (r *RAGService) StoreTransaction(ctx context.Context, userID string, transaction *model.TransactionRecord) error {
	// Create content description
	content := fmt.Sprintf("Transaction: %s, Amount: %.2f, Type: %s, Status: %s, Date: %s",
		transaction.TransactionID,
		transaction.Amount,
		transaction.Type,
		transaction.Status,
		transaction.Timestamp.Format("2006-01-02"),
	)

	docID := fmt.Sprintf("txn_%s_%s", userID, transaction.TransactionID)
	
	doc := &Document{
		ID:     docID,
		Content: content,
		Metadata: map[string]interface{}{
			"transaction_id": transaction.TransactionID,
			"amount":          transaction.Amount,
			"type":           transaction.Type,
			"status":         transaction.Status,
			"timestamp":      transaction.Timestamp,
		},
		Type:      "transaction",
		UserID:    userID,
		Timestamp: transaction.Timestamp,
	}

	// Generate embedding
	embedding, err := r.generateEmbedding(ctx, content)
	if err != nil {
		embedding = r.generateTFIDFEmbedding(content)
	}
	doc.Embedding = embedding

	r.mu.Lock()
	r.documents[docID] = doc
	r.mu.Unlock()

	return nil
}

// StoreUserContext stores user account context
func (r *RAGService) StoreUserContext(ctx context.Context, userID string, profile *model.UserProfile) error {
	content := fmt.Sprintf("User Profile: Account Age: %d days, Balance: %.2f, Account Type: %s, KYC Status: %s, Credit Score: %d",
		profile.AccountAge,
		profile.TotalBalance,
		profile.AccountType,
		profile.KYCStatus,
		profile.CreditScore,
	)

	docID := fmt.Sprintf("profile_%s", userID)
	
	doc := &Document{
		ID:      docID,
		Content: content,
		Metadata: map[string]interface{}{
			"account_age":       profile.AccountAge,
			"total_balance":     profile.TotalBalance,
			"account_type":      profile.AccountType,
			"kyc_status":        profile.KYCStatus,
			"credit_score":      profile.CreditScore,
			"transaction_count": profile.TransactionCount,
		},
		Type:      "account",
		UserID:    userID,
		Timestamp: time.Now(),
	}

	embedding, err := r.generateEmbedding(ctx, content)
	if err != nil {
		embedding = r.generateTFIDFEmbedding(content)
	}
	doc.Embedding = embedding

	r.mu.Lock()
	r.documents[docID] = doc
	r.mu.Unlock()

	return nil
}

// RetrieveRelevantContext retrieves relevant context based on query
func (r *RAGService) RetrieveRelevantContext(ctx context.Context, userID, query string, limit int) ([]*Document, error) {
	if limit <= 0 {
		limit = 5
	}

	// Generate embedding for query
	queryEmbedding, err := r.generateEmbedding(ctx, query)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate query embedding, using TF-IDF")
		queryEmbedding = r.generateTFIDFEmbedding(query)
	}

	// Get all documents for this user
	r.mu.RLock()
	userDocs := make([]*Document, 0)
	for _, doc := range r.documents {
		if doc.UserID == userID {
			userDocs = append(userDocs, doc)
		}
	}
	r.mu.RUnlock()

	// Calculate similarity scores
	for _, doc := range userDocs {
		doc.Score = r.cosineSimilarity(queryEmbedding, doc.Embedding)
	}

	// Sort by score (highest first) and return top N
	topDocs := r.topKDocuments(userDocs, limit)

	log.Debug().
		Str("user_id", userID).
		Int("retrieved", len(topDocs)).
		Msg("Retrieved relevant context from RAG")

	return topDocs, nil
}

// BuildRAGPrompt builds a prompt augmented with retrieved context
func (r *RAGService) BuildRAGPrompt(ctx context.Context, userID, query string, basePrompt string) (string, error) {
	// Retrieve relevant context
	contextDocs, err := r.RetrieveRelevantContext(ctx, userID, query, 5)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to retrieve context, using base prompt only")
		return basePrompt, nil
	}

	if len(contextDocs) == 0 {
		return basePrompt, nil
	}

	// Build context section
	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n\n**Relevant Context from Previous Conversations and User Data:**\n\n")

	for i, doc := range contextDocs {
		contextBuilder.WriteString(fmt.Sprintf("%d. [%s] %s", i+1, doc.Type, doc.Content))
		if doc.Score > 0 {
			contextBuilder.WriteString(fmt.Sprintf(" (Relevance: %.2f)", doc.Score))
		}
		contextBuilder.WriteString("\n")
	}

	contextBuilder.WriteString("\n**Use this context to provide more personalized and accurate responses.**\n")

	// Augment base prompt with context
	augmentedPrompt := basePrompt + contextBuilder.String()

	return augmentedPrompt, nil
}

// generateEmbedding generates embedding using LLM service (Ollama embeddings)
func (r *RAGService) generateEmbedding(ctx context.Context, text string) ([]float64, error) {
	// Check cache first
	r.embeddingMu.RLock()
	if cached, exists := r.embeddingCache[text]; exists {
		r.embeddingMu.RUnlock()
		return cached, nil
	}
	r.embeddingMu.RUnlock()

	// Try to use Ollama embeddings API if available
	// For now, fallback to TF-IDF
	embedding := r.generateTFIDFEmbedding(text)

	// Cache it
	r.embeddingMu.Lock()
	r.embeddingCache[text] = embedding
	r.embeddingMu.Unlock()

	return embedding, nil
}

// generateTFIDFEmbedding generates a simple TF-IDF based embedding (fallback)
func (r *RAGService) generateTFIDFEmbedding(text string) []float64 {
	// Simple word frequency vector (can be enhanced with proper TF-IDF)
	words := strings.Fields(strings.ToLower(text))
	wordFreq := make(map[string]float64)
	
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) > 0 {
			wordFreq[word]++
		}
	}

	// Normalize to unit vector
	total := 0.0
	for _, freq := range wordFreq {
		total += freq * freq
	}
	norm := math.Sqrt(total)
	if norm == 0 {
		norm = 1
	}

	// Create fixed-size vector (128 dimensions)
	embedding := make([]float64, 128)
	idx := 0
	for _, freq := range wordFreq {
		if idx >= 128 {
			break
		}
		embedding[idx] = freq / norm
		idx++
	}

	return embedding
}

// cosineSimilarity calculates cosine similarity between two vectors
func (r *RAGService) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// topKDocuments returns top K documents by score
func (r *RAGService) topKDocuments(docs []*Document, k int) []*Document {
	if len(docs) <= k {
		// Sort all
		for i := 0; i < len(docs)-1; i++ {
			for j := i + 1; j < len(docs); j++ {
				if docs[i].Score < docs[j].Score {
					docs[i], docs[j] = docs[j], docs[i]
				}
			}
		}
		return docs
	}

	// Simple selection sort for top K
	result := make([]*Document, k)
	for i := 0; i < k; i++ {
		maxIdx := i
		for j := i + 1; j < len(docs); j++ {
			if docs[j].Score > docs[maxIdx].Score {
				maxIdx = j
			}
		}
		result[i] = docs[maxIdx]
		docs[i], docs[maxIdx] = docs[maxIdx], docs[i]
	}

	return result
}

// GetUserContextSummary returns a summary of user's context for prompt augmentation
func (r *RAGService) GetUserContextSummary(ctx context.Context, userID string) (string, error) {
	// Retrieve recent conversations and transactions
	recentDocs, err := r.RetrieveRelevantContext(ctx, userID, "recent activity", 10)
	if err != nil {
		return "", err
	}

	if len(recentDocs) == 0 {
		return "", nil
	}

	var summaryBuilder strings.Builder
	summaryBuilder.WriteString("**User Context Summary:**\n\n")

	// Group by type
	conversations := []*Document{}
	transactions := []*Document{}
	accountInfo := []*Document{}

	for _, doc := range recentDocs {
		switch doc.Type {
		case "conversation":
			conversations = append(conversations, doc)
		case "transaction":
			transactions = append(transactions, doc)
		case "account":
			accountInfo = append(accountInfo, doc)
		}
	}

	// Add account info
	if len(accountInfo) > 0 {
		summaryBuilder.WriteString("Account Information:\n")
		for _, doc := range accountInfo {
			summaryBuilder.WriteString(fmt.Sprintf("- %s\n", doc.Content))
		}
		summaryBuilder.WriteString("\n")
	}

	// Add recent transactions
	if len(transactions) > 0 {
		summaryBuilder.WriteString("Recent Transactions:\n")
		for i, doc := range transactions {
			if i >= 5 { // Limit to 5 most recent
				break
			}
			summaryBuilder.WriteString(fmt.Sprintf("- %s\n", doc.Content))
		}
		summaryBuilder.WriteString("\n")
	}

	// Add recent conversation context
	if len(conversations) > 0 {
		summaryBuilder.WriteString("Recent Conversation Context:\n")
		for i, doc := range conversations {
			if i >= 3 { // Limit to 3 most recent
				break
			}
			summaryBuilder.WriteString(fmt.Sprintf("- %s\n", doc.Content))
		}
	}

	return summaryBuilder.String(), nil
}

// ClearUserContext clears all context for a user (for testing/privacy)
func (r *RAGService) ClearUserContext(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	for id, doc := range r.documents {
		if doc.UserID == userID {
			delete(r.documents, id)
			count++
		}
	}

	log.Info().
		Str("user_id", userID).
		Int("deleted", count).
		Msg("Cleared user context from RAG")
}

