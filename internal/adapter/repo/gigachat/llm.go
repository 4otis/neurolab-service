package gigachat

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/4otis/neurolab-service/internal/port/repo"
	gigachat "github.com/saintbyte/gigachat_api"
)

const (
	maxRetries = 3
	retryDelay = 2 * time.Second
	jsonStart  = "### JSON_START ###"
	jsonEnd    = "### JSON_END ###"
)

var jsonRegex = regexp.MustCompile(`(?s)` + regexp.QuoteMeta(jsonStart) + `\s*(\{.*?\})\s*` + regexp.QuoteMeta(jsonEnd))
var _ repo.LLMRepo = (*LLMRepo)(nil)

type LLMRepo struct {
	client    *gigachat.Gigachat
	histories map[string][]gigachat.MessageRequest
	mu        sync.RWMutex
}

// TODO: параметр - ключ, передать в клиент
func NewLLMRepo(secret string) *LLMRepo {
	client := gigachat.NewGigachat()
	client.AuthData = secret
	return &LLMRepo{
		client:    client,
		histories: make(map[string][]gigachat.MessageRequest),
	}
}

func (r *LLMRepo) Generate(ctx context.Context, sysP, usrP string) (string, error) { // (*repo.GenerateResponse, error) {
	// if req.Prompt == "" {
	// 	return nil, fmt.Errorf("prompt required")
	// }

	// sessionID := req.Session
	// if sessionID == "" {
	// 	sessionID = r.newSessionID()
	// }

	// history := r.getHistory(sessionID)

	history := make([]gigachat.MessageRequest, 0, 2)
	history = append(history, gigachat.MessageRequest{
		Role:    gigachat.GigaChatRoleSystem,
		Content: sysP,
	})
	history = append(history, gigachat.MessageRequest{
		Role:    gigachat.GigaChatRoleUser,
		Content: usrP,
	})

	var resp string
	for attempt := 1; attempt <= maxRetries; attempt++ {
		var err error
		resp, err = r.client.ChatCompletions(history)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			return "", fmt.Errorf("chat completions failed: %w", err)
		}

		// 	parsedJSON, valid := r.tryParseJSON(resp)
		// 	result := &repo.GenerateResponse{
		// 		Content:   resp,
		// 		SessionID: sessionID,
		// 		ValidJSON: valid,
		// 	}

		// 	if valid && parsedJSON != "" {
		// 		result.Content = parsedJSON
		// 	}

		// 	r.addToHistory(sessionID, "assistant", resp)

		return resp, nil
	}

	// return nil, fmt.Errorf("all retries exhausted")

	return "	", fmt.Errorf("all retries exhausted")
}

// func (r *LLMRepo) tryParseJSON(raw string) (string, bool) {
// 	matches := jsonRegex.FindStringSubmatch(raw)
// 	if len(matches) < 2 {
// 		return "", false
// 	}

// 	jsonStr := strings.TrimSpace(matches[1])
// 	var js map[string]interface{}
// 	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
// 		return "", false
// 	}

// 	return jsonStr, true
// }

// func (r *LLMRepo) getHistory(sessionID string) []gigachat.MessageRequest {
// 	r.mu.RLock()
// 	defer r.mu.RUnlock()
// 	if history, exists := r.histories[sessionID]; exists {
// 		return append([]gigachat.MessageRequest{}, history...)
// 	}
// 	return []gigachat.MessageRequest{}
// }

// func (r *LLMRepo) addToHistory(sessionID, role, content string) {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	if r.histories[sessionID] == nil {
// 		r.histories[sessionID] = []gigachat.MessageRequest{}
// 	}

// 	r.histories[sessionID] = append(r.histories[sessionID], gigachat.MessageRequest{
// 		Role:    role,
// 		Content: content,
// 	})
// }

// func (r *LLMRepo) newSessionID() string {
// 	return fmt.Sprintf("session_%d", time.Now().UnixNano())
// }
