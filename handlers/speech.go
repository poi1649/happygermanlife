package handlers

import (
	"awesomeProject2/models"
	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
	"context"
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
					// 오디오 채널 수 (mono)
					AudioChannelCount: 1,
					// 다른 인코딩 포맷도 지원
					// Encoding: speechpb.RecognitionConfig_WEBM_OPUS, // WEBM_OPUS 인코딩 사용 시
				},
				InterimResults: true,
			},
		},
	}); err != nil {
		log.Printf("[ERROR] Speech 설정 전송 실패: %v", err)
		return
	}

	log.Printf("[INFO] Google Speech API 설정 완료 - 인코딩: LINEAR16, 샘플 레이트: 16000Hz, 언어: de-DE")

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
			log.Printf("[INFO] WebSocket 수신 메시지 타입: %d (1=텍스트, 2=바이너리), 데이터 길이: %d 바이트", messageType, len(data))

			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[ERROR] WebSocket 에러: %v", err)
				} else {
					log.Printf("[INFO] WebSocket 연결 종료: %v", err)
				}
				// When connection closes, send the final transcription
				transcriptionChan <- finalTranscription
				return
			}

			if messageType == websocket.CloseMessage {
				log.Printf("[INFO] WebSocket 클라이언트에 의해 정상 종료됨")
				transcriptionChan <- finalTranscription
				return
			}

			if messageType == websocket.TextMessage {
				log.Printf("[INFO] 텍스트 메시지 수신: %s", string(data))
				// 텍스트 메시지 처리 (필요시)
				continue
			}

			// Only process binary messages (audio data)
			if messageType == websocket.BinaryMessage {
				log.Printf("[INFO] 오디오 바이너리 데이터 수신 - 크기: %d 바이트", len(data))

				// 첫 16바이트를 로깅 (디버깅용)
				if len(data) > 16 {
					log.Printf("[DEBUG] 오디오 데이터 시작 바이트: % x", data[:16])
				}
				// 클라이언트가 전송한 바이너리 데이터를 Google Speech API로 전송
				if err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
						AudioContent: data,
					},
				}); err != nil {
					log.Printf("[ERROR] 오디오 데이터 전송 실패: %v", err)
					continue
				}

				log.Printf("[INFO] 오디오 데이터 청크 (%d 바이트) 전송 완료", len(data))

				// Google Speech API로부터 변환 결과 수신
				resp, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						log.Printf("[INFO] 스트림 종료 (EOF)")
						break
					}
					log.Printf("[ERROR] 응답 수신 실패: %v", err)
					continue
				}

				log.Printf("[INFO] Google Speech API 응답 수신 - 결과 수: %d", len(resp.Results))

				// 변환 결과 처리
				for _, result := range resp.Results {
					if len(result.Alternatives) > 0 {
						transcript := result.Alternatives[0].Transcript
						confidence := result.Alternatives[0].Confidence

						log.Printf("[INFO] 변환 결과: '%s', 확실성: %.2f, 최종여부: %t",
							transcript, confidence, result.IsFinal)

						// 중간 결과를 클라이언트에 전송
						response := map[string]interface{}{
							"transcript": transcript,
							"final":      result.IsFinal,
							"confidence": confidence,
						}

						if err := conn.WriteJSON(response); err != nil {
							log.Printf("[ERROR] WebSocket 메시지 전송 실패: %v", err)
						} else {
							log.Printf("[INFO] 클라이언트에 결과 전송 완료")
						}

						// 최종 결과인 경우 저장
						if result.IsFinal {
							finalTranscription = transcript
							log.Printf("[INFO] 최종 텍스트 업데이트: %s", finalTranscription)
						}
					} else {
						log.Printf("[WARN] 변환 결과에 대안이 없음")
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
