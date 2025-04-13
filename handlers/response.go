package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"awesomeProject2/models"
)

// GPT4ResponseFormat represents the expected response from OpenAI
type GPT4ResponseFormat struct {
	KoreanTranslation string   `json:"korean_translation"`
	Responses         []Response `json:"responses"`
}

// Response represents a suggested response with its translation
type Response struct {
	German string `json:"german"`
	Korean string `json:"korean"`
}

// HandleGenerateResponse handles requests to generate responses using GPT-4o
func HandleGenerateResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var requestBody struct {
		Username string `json:"username"`
		Context  struct {
			Service string `json:"service"`
			Issue   string `json:"issue"`
		} `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user conversations
	conversations, exists := GetConversations(requestBody.Username)
	if !exists || len(conversations) == 0 {
		http.Error(w, "No conversations found for this user", http.StatusNotFound)
		return
	}

	// Extract latest question and previous conversation
	latestIdx := len(conversations) - 1
	latestQuestion := conversations[latestIdx].Question
	
	var previousQuestion, previousAnswer string
	if latestIdx > 0 {
		previousQuestion = conversations[latestIdx-1].Question
		previousAnswer = conversations[latestIdx-1].Answer
	}

	// Construct prompt for GPT-4o
	prompt := constructGPT4oPrompt(
		requestBody.Context.Service,
		requestBody.Context.Issue,
		latestQuestion,
		previousQuestion,
		previousAnswer,
	)

	// Call OpenAI API
	response, err := callOpenAIAPI(prompt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calling OpenAI API: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the response directly to client
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// constructGPT4oPrompt creates a prompt for GPT-4o
func constructGPT4oPrompt(service, issue, latestQuestion, previousQuestion, previousAnswer string) string {
	previousConversationContext := ""
	if previousQuestion != "" {
		previousConversationContext = fmt.Sprintf(
			"Previous user question: %s\nPrevious response: %s\n\n",
			previousQuestion,
			previousAnswer,
		)
	}

	return fmt.Sprintf(
		`You are a customer service assistant for German customers. 
Context: User is contacting about %s service regarding %s issue.

%sLatest user question: %s

Please provide:
1. Korean translation of the latest user's question
2. Two recommended responses in German for a customer service agent to reply with
3. Korean translations of each of those recommended responses

Format your response as a JSON object with the following structure:
{
  "korean_translation": "Korean translation of user's question",
  "responses": [
    {
      "german": "First recommended response in German",
      "korean": "Korean translation of first response"
    },
    {
      "german": "Second recommended response in German",
      "korean": "Korean translation of second response"
    }
  ]
}`,
		service,
		issue,
		previousConversationContext,
		latestQuestion,
	)
}

// callOpenAIAPI sends a request to the OpenAI API and returns the response
func callOpenAIAPI(prompt string) ([]byte, error) {
	url := "https://api.openai.com/v1/chat/completions"
	
	// Create request body
	requestBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a customer service assistant that helps with German and Korean languages.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}
	
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, err
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+models.GetOpenAIAPIKey())
	
	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}
	
	// Parse OpenAI response
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	
	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		return nil, err
	}
	
	if len(openAIResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}
	
	// Extract the content (should be a JSON string)
	responseContent := openAIResponse.Choices[0].Message.Content
	
	// Parse the content to verify it's valid JSON
	var parsedResponse GPT4ResponseFormat
	if err := json.Unmarshal([]byte(responseContent), &parsedResponse); err != nil {
		// If it's not valid JSON, return a formatted JSON response
		log.Printf("OpenAI didn't return valid JSON. Error: %v", err)
		log.Printf("Response content: %s", responseContent)
		
		// Try to extract information and create a valid JSON response
		// This is a fallback in case GPT doesn't return proper JSON
		return createFallbackResponse(responseContent)
	}
	
	// If it's valid JSON, return it
	return []byte(responseContent), nil
}

// createFallbackResponse attempts to create a valid JSON response when GPT doesn't return proper JSON
func createFallbackResponse(content string) ([]byte, error) {
	// Simple fallback - in a real application, you might want to do more sophisticated parsing
	fallback := GPT4ResponseFormat{
		KoreanTranslation: "Translation could not be parsed",
		Responses: []Response{
			{
				German: "Es tut uns leid, wir konnten Ihre Anfrage nicht richtig verarbeiten. Könnten Sie bitte Ihre Frage wiederholen?",
				Korean: "죄송합니다. 귀하의 요청을 제대로 처리할 수 없었습니다. 질문을 반복해 주시겠습니까?",
			},
			{
				German: "Entschuldigung für die Unannehmlichkeiten. Bitte versuchen Sie es erneut oder kontaktieren Sie uns später.",
				Korean: "불편을 끼쳐 드려 죄송합니다. 다시 시도하시거나 나중에 문의해 주세요.",
			},
		},
	}
	
	return json.Marshal(fallback)
}
