# RAG Implementation - Changes and Flow Documentation

## 📋 Overview

This document details all changes made to implement the RAG (Retrieval-Augmented Generation) system and how it transformed the conversation flow from a simple chat to a context-aware, memory-enabled banking assistant.

---

## 🔄 Flow Changes: Before vs After

### **BEFORE RAG Implementation**

```
User Input
    ↓
Intent Parser
    ↓
Context Enricher (User Profile, Risk, etc.)
    ↓
MCP Client → Agent
    ↓
LLM Service (with conversation history only)
    ↓
Response
```

**Limitations:**
- ❌ No memory of past conversations beyond current session
- ❌ No access to transaction history context
- ❌ No personalized responses based on user's account
- ❌ Context lost when session ends
- ❌ Generic responses without user-specific context

---

### **AFTER RAG Implementation**

```
User Input
    ↓
Intent Parser
    ↓
┌─────────────────────────────────────┐
│  RAG Service (NEW)                   │
│  - Store user input                  │
│  - Retrieve relevant context         │
│  - Build context-augmented prompt    │
└─────────────────────────────────────┘
    ↓
Context Enricher (User Profile, Risk, etc.)
    ↓
┌─────────────────────────────────────┐
│  RAG Service                        │
│  - Store user profile                │
│  - Store transaction history         │
└─────────────────────────────────────┘
    ↓
MCP Client → Agent
    ↓
┌─────────────────────────────────────┐
│  LLM Service                        │
│  - Prompt augmented with RAG context│
│  - Includes:                         │
│    • Past conversations              │
│    • Transaction history            │
│    • Account information            │
│    • Recent activity                │
└─────────────────────────────────────┘
    ↓
Response
    ↓
┌─────────────────────────────────────┐
│  RAG Service                        │
│  - Store assistant response          │
│  - Update context for next query     │
└─────────────────────────────────────┘
```

**Benefits:**
- ✅ Persistent memory across sessions
- ✅ Access to transaction history
- ✅ Personalized responses based on user context
- ✅ Context-aware conversations
- ✅ Semantic search for relevant information

---

## 📝 Detailed Changes

### 1. **New Files Created**

#### `ai-skin-orchestrator/internal/service/rag_service.go`
**Purpose:** Core RAG service for vector storage and retrieval

**Key Components:**
- `Document` struct: Represents stored context (conversations, transactions, account info)
- `RAGService` struct: Manages vector storage and retrieval
- Methods:
  - `StoreConversation()`: Stores user/assistant messages with embeddings
  - `StoreTransaction()`: Stores transaction records
  - `StoreUserContext()`: Stores account/profile information
  - `RetrieveRelevantContext()`: Semantic search for relevant context
  - `BuildRAGPrompt()`: Augments prompts with retrieved context
  - `GetUserContextSummary()`: Creates context summary for prompts

**Embedding Strategy:**
- Primary: TF-IDF based embeddings (128-dimensional vectors)
- Extensible: Can integrate Ollama embeddings API
- Similarity: Cosine similarity for retrieval

---

### 2. **Modified Files**

#### `ai-skin-orchestrator/internal/service/orchestrator.go`

**Changes:**

1. **Added RAG Service to Orchestrator**
   ```go
   type Orchestrator struct {
       // ... existing fields
       ragService *RAGService  // NEW
   }
   ```

2. **Updated Constructor**
   ```go
   func NewOrchestrator(
       // ... existing params
       ragService *RAGService,  // NEW parameter
   ) *Orchestrator
   ```

3. **Modified `ProcessRequest()` Flow**
   - **Step 2 (NEW):** Store user input in RAG
   ```go
   if o.ragService != nil {
       o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "user", req.Input)
   }
   ```
   
   - **Step 3 (ENHANCED):** Store enriched context in RAG
   ```go
   // Store user profile and transactions in RAG
   if o.ragService != nil {
       o.ragService.StoreUserContext(ctx, req.UserID, &enrichedContext.UserProfile)
       for _, txn := range enrichedContext.TransactionHistory {
           o.ragService.StoreTransaction(ctx, req.UserID, &txn)
       }
   }
   ```
   
   - **Step 6 (NEW):** Store agent response in RAG
   ```go
   if o.ragService != nil && agentResponse.Explanation != "" {
       o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "assistant", agentResponse.Explanation)
   }
   ```

4. **Enhanced `handleConversationalQuery()` Method**
   - **NEW:** Store user input in RAG
   - **NEW:** Retrieve user context summary
   - **NEW:** Build RAG-augmented prompt
   - **ENHANCED:** LLM call with full context
   - **NEW:** Store assistant response in RAG

   **Before:**
   ```go
   // Simple history-based call
   response, err = o.llmService.CallLLMWithHistory(ctx, req.Input, conversationHistory)
   ```

   **After:**
   ```go
   // Get user context summary
   contextSummary, err := o.ragService.GetUserContextSummary(ctx, req.UserID)
   
   // Build RAG-augmented prompt
   ragPrompt, err := o.ragService.BuildRAGPrompt(ctx, req.UserID, req.Input, basePrompt)
   
   // Full prompt with RAG context
   fullPrompt := bankingSystemPrompt + ragPrompt + conversationHistory
   response, err = o.llmService.CallLLM(ctx, fullPrompt)
   
   // Store response in RAG
   o.ragService.StoreConversation(ctx, req.UserID, req.SessionID, "assistant", response)
   ```

---

#### `ai-skin-orchestrator/cmd/server/main.go`

**Changes:**

1. **Added RAG Service Initialization**
   ```go
   // Initialize RAG service for context-aware conversations
   ragService := service.NewRAGService(llmService)
   ```

2. **Updated Orchestrator Creation**
   ```go
   orchestrator := service.NewOrchestrator(
       // ... existing services
       ragService,  // NEW
   )
   ```

---

## 🔀 Flow Comparison: Detailed Steps

### **Conversational Query Flow**

#### BEFORE:
```
1. User: "What was my last transaction?"
2. Intent Parser → IntentConversational
3. Get conversation history (last 10 messages from session)
4. Call LLM with history
5. Response: Generic answer (no transaction context)
```

#### AFTER:
```
1. User: "What was my last transaction?"
2. Intent Parser → IntentConversational
3. RAG: Store user input
4. RAG: Retrieve relevant context:
   - Past conversations about transactions
   - Recent transaction records
   - Account information
5. RAG: Build augmented prompt with:
   - User context summary
   - Recent transactions
   - Past conversation context
6. Get conversation history (last 10 messages)
7. Call LLM with FULL context (RAG + history)
8. Response: Context-aware answer with actual transaction details
9. RAG: Store assistant response
```

---

### **Banking Operation Flow (e.g., Check Balance)**

#### BEFORE:
```
1. User: "Check my balance"
2. Intent Parser → CHECK_BALANCE
3. Context Enricher → User profile, risk
4. MCP Client → Banking Agent
5. Agent returns balance
6. Response merged and returned
```

#### AFTER:
```
1. User: "Check my balance"
2. Intent Parser → CHECK_BALANCE
3. RAG: Store user input
4. Context Enricher → User profile, risk, transactions
5. RAG: Store user profile and transactions
6. MCP Client → Banking Agent
7. Agent returns balance
8. RAG: Store agent response
9. Response merged and returned
```

**Note:** For banking operations, RAG stores context for future queries. For example:
- Next query: "How much did I spend last month?" 
- RAG can retrieve: Past balance checks, transaction history, account info

---

## 📊 Data Flow: What Gets Stored

### **1. Conversation Storage**
```go
Document {
    ID: "conv_session123_1234567890",
    Content: "User: What was my last transaction?",
    Type: "conversation",
    UserID: "U10001",
    SessionID: "session123",
    Metadata: {
        "role": "user",
        "user_id": "U10001",
        "session_id": "session123"
    },
    Embedding: [0.12, 0.34, ...] // 128-dim vector
}
```

### **2. Transaction Storage**
```go
Document {
    ID: "txn_U10001_TXN123",
    Content: "Transaction: TXN123, Amount: 5000.00, Type: NEFT, Status: SUCCESS",
    Type: "transaction",
    UserID: "U10001",
    Metadata: {
        "transaction_id": "TXN123",
        "amount": 5000.00,
        "type": "NEFT",
        "status": "SUCCESS"
    },
    Embedding: [0.45, 0.67, ...]
}
```

### **3. Account Context Storage**
```go
Document {
    ID: "profile_U10001",
    Content: "User Profile: Account Age: 365 days, Balance: 150000.00, Account Type: SAVINGS",
    Type: "account",
    UserID: "U10001",
    Metadata: {
        "account_age": 365,
        "total_balance": 150000.00,
        "account_type": "SAVINGS"
    },
    Embedding: [0.23, 0.56, ...]
}
```

---

## 🔍 Context Retrieval Process

### **When User Asks: "What was my last transaction?"**

1. **Query Embedding Generation**
   ```go
   queryEmbedding = generateTFIDFEmbedding("What was my last transaction?")
   // Result: [0.15, 0.32, 0.08, ...] (128 dimensions)
   ```

2. **Similarity Calculation**
   ```go
   For each document in user's context:
       similarity = cosineSimilarity(queryEmbedding, doc.Embedding)
       doc.Score = similarity
   ```

3. **Top-K Retrieval**
   ```go
   topDocs = topKDocuments(allUserDocs, k=5)
   // Returns 5 most relevant documents
   ```

4. **Context Augmentation**
   ```
   Retrieved Context:
   1. [transaction] Transaction: TXN123, Amount: 5000.00, Type: NEFT (Relevance: 0.89)
   2. [conversation] User: Show me my transactions (Relevance: 0.76)
   3. [transaction] Transaction: TXN122, Amount: 10000.00, Type: IMPS (Relevance: 0.71)
   4. [account] User Profile: Balance: 150000.00 (Relevance: 0.65)
   5. [conversation] Assistant: Your balance is ₹1,50,000 (Relevance: 0.58)
   ```

5. **Prompt Building**
   ```
   Base Prompt: "You are a banking assistant..."
   
   + Relevant Context:
   - Recent Transactions: TXN123 (₹5,000), TXN122 (₹10,000)
   - Past conversations about transactions
   - Account balance: ₹1,50,000
   
   + Conversation History:
   User: Show me my transactions
   Assistant: Here are your recent transactions...
   
   + Current Query:
   User: What was my last transaction?
   Assistant:
   ```

---

## 🎯 Key Improvements

### **1. Memory & Context**
- **Before:** Only remembers current session (last 20 messages)
- **After:** Remembers across sessions, transactions, account info

### **2. Personalization**
- **Before:** Generic responses
- **After:** Responses based on user's actual data

### **3. Semantic Understanding**
- **Before:** Keyword matching only
- **After:** Semantic similarity search

### **4. Context Relevance**
- **Before:** All history included (may be irrelevant)
- **After:** Only relevant context retrieved and included

### **5. Data Persistence**
- **Before:** Context lost when session ends
- **After:** Context persists in RAG storage

---

## 🔧 Technical Details

### **Embedding Strategy**
- **Current:** TF-IDF based (128-dimensional vectors)
- **Future:** Can integrate Ollama embeddings API for better semantic understanding

### **Similarity Metric**
- **Method:** Cosine Similarity
- **Formula:** `cos(θ) = (A · B) / (||A|| × ||B||)`
- **Range:** 0.0 (no similarity) to 1.0 (identical)

### **Storage**
- **Current:** In-memory map (for development)
- **Future:** Can migrate to:
  - ChromaDB
  - Pinecone
  - PostgreSQL with pgvector
  - Redis with vector search

### **Retrieval Strategy**
- **Top-K:** Returns top 5 most relevant documents
- **Filtering:** User-scoped (only user's own context)
- **Grouping:** By document type (conversation, transaction, account)

---

## 📈 Example Scenarios

### **Scenario 1: Follow-up Question**

**User Session 1:**
```
User: "What's my balance?"
Assistant: "Your account balance is ₹1,50,000"
[RAG stores: conversation, account info]
```

**User Session 2 (Next Day):**
```
User: "How much did I have yesterday?"
[RAG retrieves: Previous balance conversation]
Assistant: "Based on our previous conversation, your balance was ₹1,50,000 yesterday"
```

### **Scenario 2: Transaction Query**

**User:**
```
User: "Show me transactions above ₹5,000"
[RAG retrieves: Recent transactions, account info]
Assistant: "Here are your transactions above ₹5,000:
- TXN123: ₹10,000 (NEFT) on 2024-01-15
- TXN122: ₹7,500 (IMPS) on 2024-01-14"
```

### **Scenario 3: Contextual Understanding**

**User:**
```
User: "What did I spend on groceries?"
[RAG retrieves: Past transactions, conversations about spending]
Assistant: "Based on your transaction history, I found:
- ₹2,500 at SuperMart on 2024-01-10
- ₹1,800 at GroceryStore on 2024-01-05
Total grocery spending this month: ₹4,300"
```

---

## 🚀 Performance Considerations

### **Storage Efficiency**
- Documents stored with embeddings (pre-computed)
- Embedding cache to avoid recomputation
- Automatic cleanup of old sessions (24-hour TTL)

### **Retrieval Speed**
- In-memory storage: O(n) similarity calculation
- Top-K selection: O(n log k) with optimization
- Future: Vector DB can provide sub-linear search

### **Memory Usage**
- Each document: ~1-2 KB (with embedding)
- 1000 conversations: ~1-2 MB
- Scalable to millions with proper vector DB

---

## 🔐 Privacy & Security

### **User Isolation**
- All context is user-scoped
- No cross-user data leakage
- User ID filtering in all queries

### **Data Retention**
- Session-based TTL (24 hours)
- Can implement user-controlled deletion
- `ClearUserContext()` method available

---

## 📝 Summary

### **What Changed:**
1. ✅ Added RAG service for vector storage
2. ✅ Integrated RAG into orchestrator flow
3. ✅ Enhanced conversational queries with context
4. ✅ Automatic storage of all conversations and transactions
5. ✅ Semantic retrieval of relevant context
6. ✅ Context-augmented prompts for LLM

### **Flow Transformation:**
- **Before:** Simple request → response
- **After:** Request → context retrieval → augmented prompt → context-aware response → storage

### **Impact:**
- 🎯 **Context-Aware:** Remembers past conversations and data
- 🎯 **Personalized:** Responses based on user's actual information
- 🎯 **Intelligent:** Semantic understanding of queries
- 🎯 **Persistent:** Memory across sessions
- 🎯 **Scalable:** Ready for production vector databases

---

## 🔮 Future Enhancements

1. **Ollama Embeddings:** Use actual embedding models for better semantics
2. **Vector Database:** Migrate to ChromaDB/Pinecone for production
3. **Document Chunking:** Handle longer conversations efficiently
4. **Context Summarization:** Compress very long histories
5. **Knowledge Base:** Add banking policies, FAQs, product information
6. **Multi-modal:** Support for images, documents in context

---

**Last Updated:** 2024-01-XX
**Version:** 1.0

