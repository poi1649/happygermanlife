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
			return true // Allow all connections for now, adjust as needed
		},
	}

	// In-memory data store for conversations
	conversationStore = make(map[string][]models.Conversation)
	storeMutex        sync.RWMutex
)

// HandleSpeechToText handles WebSocket connections for streaming audio data
func HandleSpeechToText(w http.ResponseWriter, r *http.Request) {
	// Extract username from header
	username := r.Header.Get("Username")
	if username == "" {
		http.Error(w, "Username header is required", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Create Google Cloud Speech client
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create speech client: %v", err)
		return
	}
	defer client.Close()

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
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				// When connection closes, send the final transcription
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
					log.Printf("Failed to send audio: %v", err)
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

	// Get the final transcription
	finalTranscription := <-transcriptionChan

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
		} else {
			// Create new conversation list for user
			conversationStore[username] = []models.Conversation{newConversation}
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
