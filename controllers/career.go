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
		fmt.Println("Error: Invalid request body format")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println("Received request with goals:", req.Goals)

	// Retrieve the OpenAI API key from the environment variables
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OpenAI API key not set")
		http.Error(w, "OpenAI API key not set", http.StatusInternalServerError)
		return
	}
	fmt.Println("OpenAI API key found, initializing client...")

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// Define the prompt for OpenAI
	prompt := fmt.Sprintf("Create a career strategy for someone aiming to achieve %s income with current skills in %s. Goals: %s", req.DesiredIncome, req.CurrentSkills, req.Goals)
	fmt.Println("Generated prompt for OpenAI:", prompt)

	// Make the request to OpenAI API
	response, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	// Check for errors and log details
	if err != nil {
		fmt.Printf("OpenAI API error: %v\n", err)
		http.Error(w, fmt.Sprintf("Failed to generate strategy: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("OpenAI API request successful")

	// Ensure response has content
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		fmt.Println("Error: Received empty response from OpenAI")
		http.Error(w, "Received empty response from OpenAI", http.StatusInternalServerError)
		return
	}
	fmt.Println("Received response from OpenAI:", response.Choices[0].Message.Content)

	// Extract and send the response
	strategy := response.Choices[0].Message.Content
	res := CareerStrategyResponse{Strategy: strategy}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
