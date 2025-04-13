#!/bin/bash

# Check if .env file exists, if not create from example
if [ ! -f .env ]; then
  echo "Creating .env file from .env.example"
  cp .env.example .env
  echo "Please edit .env file with your actual API keys"
  exit 1
fi

# Create credentials directory if it doesn't exist
mkdir -p credentials

# Check if credentials file exists
CREDENTIALS_FILE=$(grep GOOGLE_CREDENTIALS_FILE .env | cut -d '=' -f2)
if [ ! -f "$CREDENTIALS_FILE" ]; then
  echo "Google credentials file not found at $CREDENTIALS_FILE"
  echo "Please put your Google credentials JSON file at this location or update the GOOGLE_CREDENTIALS_FILE in .env"
  exit 1
fi

# Build and run with docker-compose
echo "Building and starting containers..."
docker-compose up --build
