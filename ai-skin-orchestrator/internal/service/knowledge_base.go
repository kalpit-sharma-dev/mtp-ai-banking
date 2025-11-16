package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// initializeKnowledgeBase initializes banking policies and FAQs
func (r *RAGService) initializeKnowledgeBase() {
	r.knowledgeMu.Lock()
	defer r.knowledgeMu.Unlock()

	// Banking Policies
	policies := []struct {
		id      string
		content string
		type_   string
	}{
		{
			id:      "policy_transfer_limits",
			content: "Fund Transfer Limits: NEFT - ₹10,00,000 per transaction, RTGS - Minimum ₹2,00,000, IMPS - ₹2,00,000 per day, UPI - ₹1,00,000 per transaction. Daily transfer limit is ₹5,00,000 across all channels.",
			type_:   "policy",
		},
		{
			id:      "policy_kyc_requirements",
			content: "KYC Requirements: All customers must complete KYC verification. Documents required: Aadhaar, PAN, Address proof. KYC must be updated every 2 years for active accounts.",
			type_:   "policy",
		},
		{
			id:      "policy_transaction_fees",
			content: "Transaction Fees: NEFT - Free for online, ₹2.50 for branch. RTGS - Free for online, ₹25 for branch. IMPS - ₹5 per transaction. UPI - Free. Statement charges: ₹50 per physical statement.",
			type_:   "policy",
		},
		{
			id:      "policy_account_closure",
			content: "Account Closure: Zero balance accounts can be closed online. Accounts with balance require branch visit. Closure charges: ₹500 if closed within 1 year, free after 1 year.",
			type_:   "policy",
		},
		{
			id:      "policy_fd_terms",
			content: "Fixed Deposit Terms: Minimum deposit ₹5,000. Interest rates vary by tenure: 7 days to 1 year - 5.5%, 1-3 years - 6.0%, 3-5 years - 6.5%, 5+ years - 7.0%. Premature withdrawal allowed with penalty.",
			type_:   "policy",
		},
	}

	// FAQs
	faqs := []struct {
		id      string
		content string
		type_   string
	}{
		{
			id:      "faq_balance_check",
			content: "How to check balance? You can check your account balance through mobile banking, net banking, ATM, or by asking me. Balance is updated in real-time for all transactions.",
			type_:   "faq",
		},
		{
			id:      "faq_transfer_time",
			content: "Transfer Processing Times: NEFT - Within 2 hours on working days, RTGS - Real-time on working days, IMPS - 24/7 instant, UPI - 24/7 instant. Transfers initiated after 6 PM are processed next working day.",
			type_:   "faq",
		},
		{
			id:      "faq_forgot_password",
			content: "Forgot Password: Use 'Forgot Password' on login page. OTP will be sent to registered mobile. You can also reset via ATM using debit card PIN. Contact customer care for assistance.",
			type_:   "faq",
		},
		{
			id:      "faq_beneficiary_add",
			content: "Adding Beneficiary: Add beneficiary through mobile/net banking. Verification required via OTP. New beneficiaries have 24-hour cooling period before first transfer. Maximum 50 beneficiaries allowed.",
			type_:   "faq",
		},
		{
			id:      "faq_statement_download",
			content: "Download Statement: Statements available for last 5 years. Download via mobile/net banking in PDF format. Email statements available on request. Physical statements can be requested from branch.",
			type_:   "faq",
		},
		{
			id:      "faq_loan_eligibility",
			content: "Loan Eligibility: Personal loan eligibility based on credit score (minimum 650), income, employment status. Home loan requires property documents and down payment. Check eligibility through loan calculator or contact branch.",
			type_:   "faq",
		},
		{
			id:      "faq_credit_score",
			content: "Credit Score: Check your CIBIL score through our app or website. Score ranges from 300-900. Score above 750 is excellent. Factors affecting score: payment history, credit utilization, credit history length.",
			type_:   "faq",
		},
	}

	// Store policies and FAQs (embeddings will be generated on first use)
	ctx := context.Background()
	for _, policy := range policies {
		doc := &Document{
			ID:        policy.id,
			Content:   policy.content,
			Type:      policy.type_,
			UserID:    "system", // System-wide knowledge
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"category": "banking_policy",
			},
		}
		// Generate embedding (will use TF-IDF if Ollama not available)
		if embedding, err := r.generateEmbedding(ctx, policy.content); err == nil {
			doc.Embedding = embedding
		}
		r.knowledgeBase[policy.id] = doc
	}

	for _, faq := range faqs {
		doc := &Document{
			ID:        faq.id,
			Content:   faq.content,
			Type:      faq.type_,
			UserID:    "system",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"category": "faq",
			},
		}
		if embedding, err := r.generateEmbedding(ctx, faq.content); err == nil {
			doc.Embedding = embedding
		}
		r.knowledgeBase[faq.id] = doc
	}

	log.Info().
		Int("policies", len(policies)).
		Int("faqs", len(faqs)).
		Msg("Initialized banking knowledge base")
}

// RetrieveKnowledgeBase retrieves relevant policies and FAQs based on query
func (r *RAGService) RetrieveKnowledgeBase(ctx context.Context, query string, limit int) ([]*Document, error) {
	if limit <= 0 {
		limit = 3
	}

	// Generate query embedding
	queryEmbedding, err := r.generateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	// Calculate similarity with all knowledge base documents
	r.knowledgeMu.RLock()
	allDocs := make([]*Document, 0, len(r.knowledgeBase))
	for _, doc := range r.knowledgeBase {
		allDocs = append(allDocs, doc)
	}
	r.knowledgeMu.RUnlock()

	// Calculate similarity scores
	for _, doc := range allDocs {
		doc.Score = r.cosineSimilarity(queryEmbedding, doc.Embedding)
	}

	// Return top K documents
	return r.topKDocuments(allDocs, limit), nil
}

