package handlers

import (
	"awesomeProject2/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// GPT4ResponseFormat represents the expected response from OpenAI
type GPT4ResponseFormat struct {
	KoreanTranslation string     `json:"korean_translation"`
	Responses         []Response `json:"responses"`
}

// Response represents a suggested response with its translation
type Response struct {
	German string `json:"german"`
	Korean string `json:"korean"`
}

// HandleGenerateResponse handles requests to generate responses using GPT-4o
func HandleGenerateResponse(w http.ResponseWriter, r *http.Request) {
	// 요청 로깅
	log.Printf("[INFO] Generate Response API 요청: %s %s, RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		log.Printf("[ERROR] 잘못된 HTTP 메서드: %s, RemoteAddr: %s", r.Method, r.RemoteAddr)
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
		log.Printf("[ERROR] 요청 본문 파싱 실패: %v, RemoteAddr: %s", err, r.RemoteAddr)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] 응답 생성 요청 - Username: %s, Service: %s, Issue: %s",
		requestBody.Username, requestBody.Context.Service, requestBody.Context.Issue)

	// Get user conversations with retry
	var conversations []models.Conversation
	var exists bool
	maxRetries := 3
	retryDelay := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		conversations, exists = GetConversations(requestBody.Username)
		if exists && len(conversations) > 0 {
			log.Printf("[INFO] 사용자 대화 기록 찾음 - Username: %s, 대화 수: %d",
				requestBody.Username, len(conversations))
			break // Found conversations, no need to retry
		}

		if i < maxRetries-1 {
			log.Printf("[WARN] 사용자 대화 기록 없음: %s, %v 후 재시도... (시도 %d/%d)",
				requestBody.Username, retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		}
	}

	if !exists || len(conversations) == 0 {
		log.Printf("[ERROR] 최대 재시도 후에도 대화 기록 없음 - Username: %s", requestBody.Username)
		http.Error(w, "No conversations found for this user", http.StatusNotFound)
		return
	}

	// Extract latest question and previous conversation
	latestIdx := len(conversations) - 1
	latestQuestion := conversations[latestIdx].Question
	log.Printf("[INFO] 최신 질문: %s", latestQuestion)

	var previousQuestion, previousAnswer string
	if latestIdx > 0 {
		previousQuestion = conversations[latestIdx-1].Question
		previousAnswer = conversations[latestIdx-1].Answer
		log.Printf("[INFO] 이전 대화 존재 - 질문: %s, 답변: %s", previousQuestion, previousAnswer)
	} else {
		log.Printf("[INFO] 이전 대화 없음")
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
	log.Printf("[INFO] OpenAI API 호출 시작 - Username: %s", requestBody.Username)
	response, err := callOpenAIAPI(prompt)
	if err != nil {
		log.Printf("[ERROR] OpenAI API 호출 실패: %v, Username: %s", err, requestBody.Username)
		http.Error(w, fmt.Sprintf("Error calling OpenAI API: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("[INFO] OpenAI API 응답 수신 완료 - Username: %s, 응답 길이: %d bytes",
		requestBody.Username, len(response))

	// Return the response directly to client
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	log.Printf("[INFO] 클라이언트에 응답 전송 완료 - Username: %s", requestBody.Username)
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
	log.Printf("[INFO] OpenAI API 요청 준비")
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
		log.Printf("[ERROR] 요청 JSON 생성 실패: %v", err)
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		log.Printf("[ERROR] HTTP 요청 생성 실패: %v", err)
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+models.GetOpenAIAPIKey())

	log.Printf("[INFO] OpenAI API 요청 시작")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] API 요청 전송 실패: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[INFO] OpenAI API 응답 수신 - 상태 코드: %d", resp.StatusCode)

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] 응답 본문 읽기 실패: %v", err)
		return nil, err
	}

	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] OpenAI API 오류 상태 코드: %d, 응답: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	log.Printf("[INFO] OpenAI API 응답 본문 크기: %d bytes", len(body))

	// Parse OpenAI response
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		log.Printf("[ERROR] OpenAI 응답 파싱 실패: %v", err)
		return nil, err
	}

	if len(openAIResponse.Choices) == 0 {
		log.Printf("[ERROR] OpenAI 응답에 선택지 없음")
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Extract the content (should be a JSON string)
	responseContent := openAIResponse.Choices[0].Message.Content
	log.Printf("[INFO] OpenAI 응답 콘텐츠 추출 - 길이: %d 문자", len(responseContent))

	// Parse the content to verify it's valid JSON
	var parsedResponse GPT4ResponseFormat
	if err := json.Unmarshal([]byte(responseContent), &parsedResponse); err != nil {
		// If it's not valid JSON, return a formatted JSON response
		log.Printf("[WARN] OpenAI가 유효한 JSON을 반환하지 않음. 오류: %v", err)
		log.Printf("[WARN] 응답 내용: %s", responseContent)

		// Try to extract information and create a valid JSON response
		// This is a fallback in case GPT doesn't return proper JSON
		log.Printf("[INFO] 대체 응답 생성 중")
		return createFallbackResponse()
	}

	log.Printf("[INFO] 유효한 JSON 응답 확인됨 - 번역: %s, 추천 응답 수: %d",
		parsedResponse.KoreanTranslation, len(parsedResponse.Responses))

	// If it's valid JSON, return it
	return []byte(responseContent), nil
}

// createFallbackResponse attempts to create a valid JSON response when GPT doesn't return proper JSON
func createFallbackResponse() ([]byte, error) {
	log.Printf("[INFO] 대체 응답 생성")
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

	result, err := json.Marshal(fallback)
	if err != nil {
		log.Printf("[ERROR] 대체 응답 생성 실패: %v", err)
		return nil, err
	}

	log.Printf("[INFO] 대체 응답 생성 완료")
	return result, nil
}
