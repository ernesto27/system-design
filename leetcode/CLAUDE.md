# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a LeetCode-style coding platform built with Go, featuring a REST API for managing coding problems, user submissions, and a planned code execution sandbox. The system uses PostgreSQL for data persistence and Docker for containerized development.

## Development Commands

### Database Setup
```bash
# Start PostgreSQL database
docker-compose up -d

# Run the application (includes auto-migration and seeding)
go run .

# The application will:
# - Connect to PostgreSQL on localhost:5432
# - Auto-migrate database schema (users, problems, submissions tables)
# - Seed initial data (5 users, 5 coding problems)
# - Start HTTP server on port 8080
```

### Testing
```bash
# Test API endpoints (requires server to be running)
./test_api.sh

# Run Go tests
go test ./...
```

### Dependencies
```bash
# Install/update dependencies
go mod tidy

# Download dependencies
go mod download
```

## Architecture

### Core Components

1. **Models** (`models.go`): Defines GORM models for User, Problem, and Submission entities with custom JSON marshaling for time formatting
2. **Database** (`database.go`): PostgreSQL connection management using GORM with configurable connection settings
3. **Handlers** (`handlers.go`): HTTP request handlers for CRUD operations on problems and submissions
4. **Router** (`router.go`): Gin-based REST API routing configuration
5. **Seeding** (`seed.go`): Initial data population for users and problems

### Database Schema

- **Users**: ID, Name, Email (unique), timestamps
- **Problems**: ID, Title, Description, Difficulty, TestCases (JSONB), timestamps  
- **Submissions**: ID, UserID, ProblemID, Code, Language, Status, timestamps

### API Endpoints

- `GET /problems` - List problems with optional pagination (`?start=0&end=10`)
- `GET /problems/:problem_id` - Get specific problem details
- `POST /problems/:problem_id/submission` - Submit code solution

### Code Execution Sandbox (Planned)

Located in `internal/code_executor/`, this module uses testcontainers-go for secure code execution:

- **Supported Languages**: Python, Node.js, Go, Java
- **Security**: Isolated containers, resource limits, automatic cleanup
- **Features**: Parallel execution, timeout handling, real-time output capture

The implementation uses Docker containers with Alpine-based images for minimal attack surface and includes comprehensive security measures like non-root execution and network isolation.

## Configuration

### Database Connection
Default configuration (can be modified in `database.go`):
- Host: localhost
- Port: 5432
- User: postgres
- Password: password
- Database: leetcode_db
- SSL Mode: disable

### Docker Compose Services
- PostgreSQL 17 with persistent volume
- Exposed on port 5432
- Auto-restart enabled

## Key Design Patterns

1. **Repository Pattern**: Database operations abstracted through GORM models
2. **Dependency Injection**: Handlers receive database instance via constructor
3. **JSON API**: Consistent JSON responses with proper HTTP status codes
4. **Custom Marshaling**: Time fields formatted as "2006-01-02 15:04:05" in JSON responses
5. **Error Handling**: Comprehensive validation and error responses for all endpoints

## Development Notes

- The codebase follows Go standard project layout with clear separation of concerns
- All database operations use GORM for type safety and SQL injection prevention  
- Test cases are stored as JSONB in PostgreSQL for flexible schema
- The submission system currently returns mock test results but is designed for real code execution integration
- Time fields in JSON responses are consistently formatted for frontend consumption