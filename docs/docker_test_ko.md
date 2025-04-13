# Docker를 이용한 로컬 테스트 가이드

이 가이드는 Docker를 사용하여 로컬 환경에서 독일어 고객 서비스 지원 플랫폼 백엔드를 테스트하는 방법을 설명합니다.

## 준비 사항

1. Docker와 Docker Compose가 설치되어 있어야 합니다.
2. OpenAI API 키와 Google Cloud 자격 증명이 필요합니다.

## 설정 단계

### 1. 환경 변수 설정

`.env` 파일을 생성하고 API 키 정보를 입력합니다:

```
# API Keys
OPENAI_API_KEY=your_openai_api_key_here
GOOGLE_CREDENTIALS_FILE=./credentials/google-credentials.json
```

필요하다면 `.env.example` 파일을 복사하여 사용할 수 있습니다:

```bash
cp .env.example .env
```

그런 다음 `.env` 파일을 편집하여 실제 API 키를 입력합니다.

### 2. Google 자격 증명 파일 준비

Google Cloud 서비스 계정에서 다운로드한 자격 증명 JSON 파일을 `credentials` 디렉토리에 `google-credentials.json`이라는 이름으로 저장합니다:

```bash
# credentials 디렉토리 생성
mkdir -p credentials

# 자격 증명 파일을 credentials 디렉토리로 복사
cp /path/to/your-google-credentials.json credentials/google-credentials.json
```

### 3. Docker 이미지 빌드 및 실행

제공된 실행 스크립트를 사용하여 Docker 이미지를 빌드하고 실행할 수 있습니다:

**Linux/macOS:**
```bash
chmod +x run.sh
./run.sh
```

**Windows:**
```
run.bat
```

또는 직접 Docker Compose 명령어를 실행할 수도 있습니다:
```bash
docker-compose up --build
```

## 테스트하기

서버가 실행되면 `http://localhost:8080`에서 접근할 수 있습니다.

### 1. 상태 확인 엔드포인트

브라우저 또는 curl을 사용하여 상태 확인 엔드포인트를 테스트합니다:

```bash
curl http://localhost:8080/health
```

정상적으로 작동하면 "Service is healthy" 메시지가 표시됩니다.

### 2. 응답 생성 엔드포인트 테스트

다음 curl 명령어를 사용하여 응답 생성 엔드포인트를 테스트합니다:

```bash
curl -X POST http://localhost:8080/api/generate-response \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_user",
    "context": {
      "service": "internet",
      "issue": "connection problem"
    }
  }'
```

### 3. 음성-텍스트 변환 테스트

WebSocket 테스트를 위해 `examples/websocket_client.js` 파일을 참조하여 클라이언트 애플리케이션을 구현할 수 있습니다.

## 문제 해결

### 로그 확인

Docker 컨테이너의 로그를 확인하려면:

```bash
docker-compose logs
```

실시간 로그를 확인하려면:

```bash
docker-compose logs -f
```

### 일반적인 문제

1. **API 키 문제**: 환경 변수가 올바르게 전달되었는지 확인하세요.
2. **Google 자격 증명 접근 문제**: 볼륨 마운트가 올바르게 구성되었는지 확인하세요.
3. **포트 충돌**: 8080 포트가 이미 사용 중인 경우 `docker-compose.yml` 파일에서 포트 매핑을 변경하세요.

### 컨테이너 재시작

문제가 발생한 경우 컨테이너를 재시작해 보세요:

```bash
docker-compose down
docker-compose up --build
```
