package handler

import (
	"encoding/json"
	"net/http"

	"github.com/4otis/neurolab-service/internal/cases"
)

type LLMHandler struct {
	useCase cases.LLMUseCase
}

func NewLLMHandler(useCase cases.LLMUseCase) *LLMHandler {
	return &LLMHandler{useCase: useCase}
}

func (h *LLMHandler) GenerateCourse(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Prompt string `json:"prompt"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	content, _ := h.useCase.GenerateCourse(r.Context(), req.Name, req.Prompt)

	json.NewEncoder(w).Encode(map[string]string{"content": content})
}
