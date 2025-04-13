@echo off

REM Check if .env file exists, if not create from example
if not exist .env (
  echo Creating .env file from .env.example
  copy .env.example .env
  echo Please edit .env file with your actual API keys
  exit /b 1
)

REM Create credentials directory if it doesn't exist
if not exist credentials mkdir credentials

REM Build and run with docker-compose
echo Building and starting containers...
docker-compose up --build
