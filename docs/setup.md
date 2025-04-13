# German Customer Service Assistant Backend

This document provides instructions for setting up and running the German Customer Service Assistant backend server.

## Prerequisites

1. Go 1.20 or later
2. Google Cloud account with Speech-to-Text API enabled
3. OpenAI account with access to GPT-4o model
4. Google Cloud service account credentials file

## Environment Setup

Set the following environment variables:

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-openai-api-key"

# Set path to your Google Cloud service account credentials file
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-credentials.json"
```

## Running the Server

Run the server with the following command:

```bash
go run main.go
```

By default, the server runs on port 8080. You can specify a different port using the `-port` flag:

```bash
go run main.go -port=8000
```

## API Endpoints

### Speech-to-Text Streaming Endpoint

- **URL**: `/api/speech`
- **Method**: WebSocket
- **Headers**:
  - `Username`: Required. The name of the user.
- **Description**: Establishes a WebSocket connection for streaming audio data to be transcribed.

### Generate Response Endpoint

- **URL**: `/api/generate-response`
- **Method**: POST
- **Headers**:
  - `Content-Type`: `application/json`
- **Request Body**:
  ```json
  {
    "username": "user123",
    "context": {
      "service": "internet",
      "issue": "connection problem"
    }
  }
  ```
- **Response**: JSON object containing:
  - Korean translation of the latest user's question
  - Two recommended responses in German
  - Korean translations of each recommended response

## Testing with cURL

Test the Generate Response endpoint:

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

## Troubleshooting

If you encounter issues with the Google Speech-to-Text API:
1. Verify your service account has the correct permissions
2. Ensure the Speech-to-Text API is enabled in your Google Cloud Console
3. Check that the `GOOGLE_APPLICATION_CREDENTIALS` environment variable is correctly set

If you encounter issues with the OpenAI API:
1. Verify your API key is valid
2. Ensure you have access to the GPT-4o model
3. Check that the `OPENAI_API_KEY` environment variable is correctly set
