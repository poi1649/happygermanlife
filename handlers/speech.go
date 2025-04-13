package handlers

import (
	"awesomeProject2/models"
	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"sync"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all connections for local testing
		},
		EnableCompression: true,
	}

	// In-memory data store for conversations
	conversationStore = make(map[string][]models.Conversation)
	storeMutex        sync.RWMutex
)

// HandleSpeechToText handles WebSocket connections for streaming audio data
func HandleSpeechToText(w http.ResponseWriter, r *http.Request) {
	// 요청 로깅
	log.Printf("[INFO] Speech-to-Text API 요청: %s %s, RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Set CORS headers for WebSocket
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract username from query parameter
	username := r.URL.Query().Get("Username")
	if username == "" {
		log.Printf("[ERROR] Username query parameter is missing. RemoteAddr: %s", r.RemoteAddr)
		http.Error(w, "Username query parameter is required", http.StatusBadRequest)
		return
	}

	log.Printf("[INFO] Speech-to-Text 연결 시작 - Username: %s", username)

	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ERROR] WebSocket 업그레이드 실패: %v, RemoteAddr: %s", err, r.RemoteAddr)
		return
	}
	defer conn.Close()

	// Create Google Cloud Speech client
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("[ERROR] Google Speech 클라이언트 생성 실패: %v, Username: %s", err, username)
		return
	}
	defer client.Close()

	log.Printf("[INFO] Google Speech 클라이언트 생성 성공 - Username: %s", username)

	// Create a speech recognition stream
	stream, err := client.StreamingRecognize(ctx)
	if err != nil {
		log.Printf("Failed to create streaming recognition: %v", err)
		return
	}

	// Configure the recognition
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:        speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz: 16000,
					LanguageCode:    "de-DE", // German language
				},
				InterimResults: true,
			},
		},
	}); err != nil {
		log.Printf("Failed to send config: %v", err)
		return
	}

	// Channel to signal when WebSocket is closed
	done := make(chan struct{})

	// Channel for storing final transcription
	transcriptionChan := make(chan string, 1)

	// Handle incoming audio data from WebSocket
	go func() {
		defer close(done)

		var finalTranscription string

		for {
			// Read message from WebSocket
			messageType, data, err := conn.ReadMessage()
			log.Printf("WebSocket received: %v", data)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				// When connection closes, send the final transcription
				transcriptionChan <- finalTranscription
				return
			}

			if messageType == websocket.CloseMessage {
				log.Printf("WebSocket closed by client")
				transcriptionChan <- finalTranscription
				return
			}

			// Only process binary messages (audio data)
			if messageType == websocket.BinaryMessage {
				// Send audio data to Speech-to-Text
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: data,
					},
				}); err != nil {
					log.Printf("Failed to send audio: %v", err.Error())
					continue
				}

				// Receive transcription results
				resp, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Printf("Failed to receive response: %v", err)
					continue
				}

				// Process results
				for _, result := range resp.Results {
					transcript := result.Alternatives[0].Transcript

					// Send intermediate results back to client
					if err := conn.WriteJSON(map[string]string{
						"transcript": transcript,
						"final":      fmt.Sprintf("%t", result.IsFinal),
					}); err != nil {
						log.Printf("Failed to write to WebSocket: %v", err)
					}

					// Update final transcription if result is final
					if result.IsFinal {
						finalTranscription = transcript
					}
				}
			}
		}
	}()

	// Wait for WebSocket to close
	<-done
	log.Printf("[INFO] WebSocket 연결 종료 - Username: %s", username)

	// Get the final transcription
	finalTranscription := <-transcriptionChan

	if finalTranscription != "" {
		log.Printf("[INFO] 최종 텍스트 변환 결과: %s, Username: %s", finalTranscription, username)
	} else {
		log.Printf("[WARN] 텍스트 변환 결과 없음 - Username: %s", username)
	}

	// Store conversation
	if finalTranscription != "" {
		storeMutex.Lock()
		defer storeMutex.Unlock()

		// Create a new conversation entry
		newConversation := models.Conversation{
			Question: finalTranscription,
			Answer:   "",
		}

		// Check if user exists in the store
		if conversations, exists := conversationStore[username]; exists {
			// Add new conversation to existing list
			conversationStore[username] = append(conversations, newConversation)
			log.Printf("[INFO] 기존 대화 목록에 질문 추가 - Username: %s, 총 대화 수: %d",
				username, len(conversationStore[username]))
		} else {
			// Create new conversation list for user
			conversationStore[username] = []models.Conversation{newConversation}
			log.Printf("[INFO] 새 대화 목록 생성 - Username: %s", username)
		}
	}
}

// GetConversations returns the conversations for a given user
func GetConversations(username string) ([]models.Conversation, bool) {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	conversations, exists := conversationStore[username]
	return conversations, exists
}
