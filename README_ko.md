# 독일어 고객 서비스 지원 플랫폼 백엔드

이 프로젝트는 독일어 고객 서비스 지원을 위한 Golang 백엔드 서버를 구현합니다. 음성-텍스트 변환과 GPT-4o를 활용한 응답 생성 기능을 제공합니다.

## 주요 기능

1. **음성-텍스트 변환 스트리밍 엔드포인트**
   - WebSocket을 통한 실시간 오디오 스트리밍 처리
   - Google Cloud Speech-to-Text API를 사용한 독일어 음성 인식
   - 사용자별 대화 내역 저장

2. **응답 생성 엔드포인트**
   - OpenAI GPT-4o를 활용한 지능형 응답 생성
   - 컨텍스트 인식 대화 처리
   - 독일어 응답 및 한국어 번역 제공

## 기술 스택

- Golang 1.20+
- Gorilla Mux 및 WebSocket
- Google Cloud Speech-to-Text API
- OpenAI GPT-4o API
- Docker 및 Docker Compose

## 시작하기

### 사전 요구사항

- Go 1.20 이상
- Docker 및 Docker Compose (컨테이너화 배포 시)
- Google Cloud 계정 및 Speech-to-Text API 접근 권한
- OpenAI API 키 (GPT-4o 접근 권한 포함)

### 로컬 환경 설정

1. 저장소 클론:
   ```bash
   git clone <repository-url>
   cd awesomeProject2
   ```

2. 의존성 설치:
   ```bash
   go mod download
   ```

3. 환경 변수 설정:
   ```bash
   export OPENAI_API_KEY="your-openai-api-key"
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-credentials.json"
   ```

4. 서버 실행:
   ```bash
   go run main.go
   ```

### Docker를 사용한 설정

1. `.env` 파일 생성:
   ```bash
   cp .env.example .env
   # .env 파일을 편집하여 실제 API 키 입력
   ```

2. Google 자격 증명 설정:
   ```bash
   mkdir -p credentials
   # Google Cloud 서비스 계정에서 다운로드한 자격 증명 파일 복사
   cp /path/to/your-credentials.json credentials/google-credentials.json
   ```

3. Docker Compose로 실행:
   ```bash
   # Linux/macOS:
   ./run.sh

   # Windows:
   run.bat
   ```

## API 엔드포인트

### 1. 음성-텍스트 변환 (WebSocket)

- **URL**: `/api/speech`
- **방식**: WebSocket
- **쿼리 파라미터**: 
  - `Username`: 사용자 식별자
- **설명**: 오디오 데이터를 실시간으로 텍스트로 변환

### 2. 응답 생성

- **URL**: `/api/generate-response`
- **방식**: POST
- **요청 본문**:
  ```json
  {
    "username": "user123",
    "context": {
      "service": "internet",
      "issue": "connection problem"
    }
  }
  ```
- **응답**: 사용자 질문 번역과 추천 응답이 포함된 JSON

### 3. 상태 확인

- **URL**: `/health`
- **방식**: GET
- **응답**: 서비스 상태 정보

## 프로젝트 구조

```
awesomeProject2/
├── docs/                      # 문서
│   ├── setup.md               # 설정 지침 (영어)
│   ├── setup_ko.md            # 설정 지침 (한국어)
│   ├── docker_test_en.md      # Docker 테스트 가이드 (영어)
│   └── docker_test_ko.md      # Docker 테스트 가이드 (한국어)
├── examples/
│   └── websocket_client.js    # 테스트용 클라이언트 예제
├── handlers/
│   ├── response.go            # 응답 생성 핸들러
│   └── speech.go              # 음성-텍스트 핸들러
├── models/
│   ├── config.go              # 구성 헬퍼
│   └── conversation.go        # 대화 데이터 모델
├── .env.example               # 환경 변수 템플릿
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod                     # Go 모듈 의존성
├── main.go                    # 메인 애플리케이션
├── run.bat                    # Windows 실행 스크립트
└── run.sh                     # Linux/macOS 실행 스크립트
```

## 사용 예시

### 응답 생성 요청 예시

```bash
curl -X POST http://localhost:8080/api/generate-response \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user123",
    "context": {
      "service": "internet",
      "issue": "connection problem"
    }
  }'
```

### 예상 응답 형식

```json
{
  "korean_translation": "인터넷 연결이 자꾸 끊어집니다. 어떻게 해결할 수 있나요?",
  "responses": [
    {
      "german": "Es tut mir leid, dass Sie Probleme mit Ihrer Internetverbindung haben. Können Sie mir bitte sagen, welchen Router-Typ Sie verwenden und wann das Problem begonnen hat?",
      "korean": "인터넷 연결 문제가 발생하여 죄송합니다. 어떤 유형의 라우터를 사용하고 계신지, 언제부터 문제가 시작되었는지 알려주시겠어요?"
    },
    {
      "german": "Ich verstehe Ihr Problem mit der Internetverbindung. Haben Sie bereits versucht, Ihren Router neu zu starten? Falls nicht, empfehle ich, den Router für etwa 30 Sekunden vom Strom zu trennen und dann wieder anzuschließen.",
      "korean": "인터넷 연결 문제에 대해 이해합니다. 이미 라우터를 재시작해 보셨나요? 그렇지 않다면, 라우터의 전원을 약 30초 동안 분리한 후 다시 연결해 보시는 것을 권장합니다."
    }
  ]
}
```

## 문제 해결

자세한 문제 해결 가이드는 `docs/setup_ko.md` 및 `docs/docker_test_ko.md` 문서를 참조하세요.

### 로그 확인

Docker 환경에서 로그를 확인하려면:
```bash
docker-compose logs -f
```

## 향후 개발 계획

1. 대화 답변 저장 기능 구현
2. 사용자 인증 추가
3. 다국어 지원 확장
4. 데이터베이스 연동으로 영구 저장소 구현
