# Local Testing with Docker

This guide explains how to test the German customer service assistance platform backend in a local environment using Docker.

## Prerequisites

1. Docker and Docker Compose must be installed.
2. You need an OpenAI API key and Google Cloud credentials.

## Setup Steps

### 1. Set Environment Variables

Create a `.env` file and enter your API key information:

```
# API Keys
OPENAI_API_KEY=your_openai_api_key_here
GOOGLE_CREDENTIALS_FILE=./credentials/google-credentials.json
```

You can copy the `.env.example` file if needed:

```bash
cp .env.example .env
```

Then edit the `.env` file to insert your actual API keys.

### 2. Prepare Google Credentials File

Save the credentials JSON file downloaded from your Google Cloud service account to the `credentials` directory as `google-credentials.json`:

```bash
# Create credentials directory
mkdir -p credentials

# Copy your credentials file to the credentials directory
cp /path/to/your-google-credentials.json credentials/google-credentials.json
```

### 3. Build and Run the Docker Image

You can build and run the Docker image using the provided run script:

**Linux/macOS:**
```bash
chmod +x run.sh
./run.sh
```

**Windows:**
```
run.bat
```

Or you can run the Docker Compose command directly:
```bash
docker-compose up --build
```

## Testing

Once the server is running, it can be accessed at `http://localhost:8080`.

### 1. Health Check Endpoint

Test the health check endpoint using a browser or curl:

```bash
curl http://localhost:8080/health
```

If working properly, it should display the message "Service is healthy".

### 2. Test the Generate Response Endpoint

Test the response generation endpoint using the following curl command:

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

### 3. Speech-to-Text Testing

For WebSocket testing, you can implement a client application by referring to the `examples/websocket_client.js` file.

## Troubleshooting

### Checking Logs

To check the Docker container logs:

```bash
docker-compose logs
```

To view logs in real-time:

```bash
docker-compose logs -f
```

### Common Issues

1. **API Key Issues**: Verify that environment variables are being passed correctly.
2. **Google Credentials Access Issues**: Make sure the volume mount is configured correctly.
3. **Port Conflicts**: If port 8080 is already in use, change the port mapping in the `docker-compose.yml` file.

### Restarting Containers

If issues occur, try restarting the containers:

```bash
docker-compose down
docker-compose up --build
```
