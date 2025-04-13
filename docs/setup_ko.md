# 독일어 고객 서비스 지원 플랫폼 백엔드 설명서

이 문서는 독일어 고객 서비스 지원 플랫폼 백엔드 서버의 설정 및 사용 방법에 대한 설명을 제공합니다.

## 사전 요구사항

1. Go 1.20 이상
2. Speech-to-Text API가 활성화된 Google Cloud 계정
3. GPT-4o 모델에 접근 권한이 있는 OpenAI 계정
4. Google Cloud 서비스 계정 자격 증명 파일

## 환경 설정

다음 환경 변수를 설정하세요:

```bash
# OpenAI API 키 설정
export OPENAI_API_KEY="your-openai-api-key"

# Google Cloud 서비스 계정 자격 증명 파일 경로 설정
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-credentials.json"
```

## 서버 실행

다음 명령어로 서버를 실행하세요:

```bash
go run main.go
```

기본적으로 서버는 8080 포트에서 실행됩니다. `-port` 플래그를 사용하여 다른 포트를 지정할 수 있습니다:

```bash
go run main.go -port=8000
```

## API 엔드포인트

### 음성-텍스트 변환 스트리밍 엔드포인트

- **URL**: `/api/speech`
- **방식**: WebSocket
- **쿼리 파라미터**:
  - `Username`: 필수. 사용자의 이름.
- **설명**: 음성 데이터를 전송하고 텍스트로 변환하기 위한 WebSocket 연결을 설정합니다.

### 응답 생성 엔드포인트

- **URL**: `/api/generate-response`
- **방식**: POST
- **헤더**:
  - `Content-Type`: `application/json`
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
- **응답**: 다음을 포함하는 JSON 객체:
  - 사용자의 최신 질문에 대한 한국어 번역
  - 독일어로 된 두 가지 추천 응답
  - 각 추천 응답에 대한 한국어 번역

## cURL을 사용한 테스트

응답 생성 엔드포인트 테스트:

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

## 시스템 구조

### 프로젝트 구조
```
awesomeProject2/
├── docs/
│   └── setup.md                # 설정 지침
├── examples/
│   └── websocket_client.js     # 테스트용 예제 클라이언트
├── handlers/
│   ├── response.go             # 응답 생성 엔드포인트 핸들러
│   └── speech.go               # 음성-텍스트 WebSocket 핸들러
├── models/
│   ├── config.go               # 구성 헬퍼
│   └── conversation.go         # 대화 데이터 모델
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod                      # 의존성
└── main.go                     # 메인 애플리케이션 진입점
```

### 핵심 구성 요소

1. **음성-텍스트 스트리밍 엔드포인트** (`handlers/speech.go`)
   - 스트리밍 오디오 데이터를 수신하는 WebSocket 핸들러 구현
   - 실시간 텍스트 변환을 위한 Google Cloud Speech-to-Text API 사용
   - 사용자 이름을 키로 하여 인메모리 맵에 변환된 텍스트 저장

2. **응답 생성 엔드포인트** (`handlers/response.go`)
   - 대화 기록을 기반으로 응답을 생성하는 요청 처리
   - 컨텍스트 정보로 프롬프트를 구성하고 OpenAI GPT-4o에 전송
   - 한국어 번역과 함께 응답 형식 지정 및 반환

3. **메인 애플리케이션** (`main.go`)
   - Gorilla Mux 라우터를 사용하여 HTTP 라우트 설정
   - 크로스 오리진 요청을 허용하도록 CORS 구성
   - 상태 확인 엔드포인트 및 환경 변수 유효성 검사 포함

### 데이터 저장 구조

인메모리 대화 저장소는 다음과 같은 맵 구조로 유지됩니다:
- 키: 사용자 이름 (문자열)
- 값: 대화 목록 (대화 객체 배열)

각 대화 객체는 다음을 포함합니다:
- `Question`: 사용자 음성에서 변환된 텍스트
- `Answer`: 응답 (초기에는 비어 있음)

## Docker를 사용한 배포

제공된 Dockerfile과 docker-compose.yml을 사용하여 애플리케이션을 컨테이너화하고 배포할 수 있습니다:

```bash
# Docker 이미지 빌드
docker build -t german-customer-service-backend .

# Docker Compose로 서비스 시작
docker-compose up -d
```

## 문제 해결

Google Speech-to-Text API 관련 문제가 발생한 경우:
1. 서비스 계정에 올바른 권한이 있는지 확인
2. Google Cloud Console에서 Speech-to-Text API가 활성화되어 있는지 확인
3. `GOOGLE_APPLICATION_CREDENTIALS` 환경 변수가 올바르게 설정되어 있는지 확인

OpenAI API 관련 문제가 발생한 경우:
1. API 키가 유효한지 확인
2. GPT-4o 모델에 접근 권한이 있는지 확인
3. `OPENAI_API_KEY` 환경 변수가 올바르게 설정되어 있는지 확인
