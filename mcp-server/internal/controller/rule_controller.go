package controller

import (
	"encoding/json"
	"net/http"

	"github.com/aibanking/mcp-server/internal/service"
)

// RuleController handles rule-related HTTP requests
type RuleController struct {
	ruleEngine *service.RuleEngine
}

// NewRuleController creates a new rule controller
func NewRuleController(ruleEngine *service.RuleEngine) *RuleController {
	return &RuleController{
		ruleEngine: ruleEngine,
	}
}

// UploadRules handles POST /rules/upload
func (rc *RuleController) UploadRules(w http.ResponseWriter, r *http.Request) {
	var rules map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	if err := rc.ruleEngine.UploadRules(rules); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to upload rules", err)
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Rules uploaded successfully",
		"count":   len(rules),
	})
}

// GetRules handles GET /rules
func (rc *RuleController) GetRules(w http.ResponseWriter, r *http.Request) {
	rules := rc.ruleEngine.GetRules()

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

