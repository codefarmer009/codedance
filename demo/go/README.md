# Go Demo Application

This is a simple Go HTTP server demo for automated deployment testing.

## Features

- Simple HTTP server with health check endpoint
- Configurable port via PORT environment variable (default: 8080)
- Docker support for containerized deployment

## Endpoints

- `/` - Returns a hello message
- `/health` - Health check endpoint

## Running Locally

```bash
go run main.go
```

## Building

```bash
go build -o demo main.go
```

## Docker Build

```bash
docker build -t codedance-demo .
docker run -p 8080:8080 codedance-demo
```

## Automated Deployment

This application is designed to be automatically compiled and deployed by CI/CD systems. The Dockerfile enables containerized deployment for cloud platforms.
