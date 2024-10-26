package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"os"
)

// CareerStrategyRequest represents the expected input for the career strategy endpoint
type CareerStrategyRequest struct {
	Goals         string `json:"goals"`
	DesiredIncome string `json:"desired_income"`
	CurrentSkills string `json:"current_skills"`
}

// CareerStrategyResponse represents the AI-generated response for a career strategy
type CareerStrategyResponse struct {
	Strategy string `json:"strategy"`
}

// GenerateCareerStrategy handles generating a career strategy using OpenAI
func GenerateCareerStrategy(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CareerStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	prompt := fmt.Sprintf("Create a career strategy for someone aiming to achieve %s income with current skills in %s. Goals: %s", req.DesiredIncome, req.CurrentSkills, req.Goals)

	response, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		http.Error(w, "Failed to generate strategy", http.StatusInternalServerError)
		return
	}

	// Prepare response
	strategy := response.Choices[0].Message.Content
	res := CareerStrategyResponse{Strategy: strategy}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
