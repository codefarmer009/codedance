# CodeDance Demo - Go Application

This is a simple Go HTTP server designed for automated deployment systems.

## Features

- HTTP server with two endpoints:
  - `/` - Returns a hello message
  - `/health` - Health check endpoint for monitoring
- Configurable port via `PORT` environment variable (default: 8080)
- Docker support for containerized deployment

## Local Development

### Prerequisites
- Go 1.21 or higher

### Running Locally

```bash
go run main.go
```

The server will start on port 8080 by default. You can customize the port:

```bash
PORT=3000 go run main.go
```

### Testing Endpoints

```bash
# Test root endpoint
curl http://localhost:8080/

# Test health endpoint
curl http://localhost:8080/health
```

## Docker Deployment

### Build Docker Image

```bash
docker build -t codedance-demo .
```

### Run Docker Container

```bash
docker run -p 8080:8080 codedance-demo
```

Or with custom port:

```bash
docker run -p 3000:3000 -e PORT=3000 codedance-demo
```

## Automated Deployment

This application is ready for automated CI/CD deployment systems. The Dockerfile uses multi-stage builds for optimized image size and security.

### Build Commands
- **Standard build**: `go build -o server .`
- **Docker build**: `docker build -t codedance-demo .`

### Environment Variables
- `PORT` - Server port (default: 8080)
