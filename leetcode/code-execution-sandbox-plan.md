# Code Execution Sandbox Module Plan (Using Testcontainers for Go)

## Overview
Create a secure Go module using **Testcontainers for Go** that accepts a programming language and code string, then executes it in a sandboxed container environment with automatic lifecycle management.

## Why Testcontainers for Go?
- **Automatic cleanup**: No resource leaks, containers destroyed after execution
- **Go-native API**: Clean, idiomatic Go implementation  
- **Built-in security**: Isolated, ephemeral containers
- **Parallel execution**: Handle multiple code executions simultaneously
- **Lifecycle hooks**: Customize behavior at each stage
- **Wait strategies**: Ensure containers are ready before execution

## Architecture

### 1. Core Components
```go
type CodeExecutor struct {
    timeout time.Duration
    resourceLimits ResourceLimits
}

type ExecutionRequest struct {
    Language string
    Code     string
    Timeout  time.Duration
}

type ExecutionResult struct {
    Stdout       string
    Stderr       string
    ExitCode     int
    ExecutionTime time.Duration
    Error        error
}
```

### 2. Language Support
- **Python**: `python:3.11-alpine` 
- **Node.js**: `node:18-alpine`
- **Go**: `golang:1.21-alpine`
- **Java**: `openjdk:17-alpine`

### 3. Implementation Pattern
```go
func (ce *CodeExecutor) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
    // Create container with Testcontainers
    container, err := testcontainers.Run(ctx,
        ce.getImageForLanguage(req.Language),
        testcontainers.WithResourceLimit(...),
        testcontainers.WithWaitStrategy(wait.ForLog("ready")),
        testcontainers.WithLifecycleHooks(...),
    )
    defer container.Terminate(ctx) // Automatic cleanup
    
    // Execute code inside container
    return ce.executeCode(ctx, container, req)
}
```

## Security Features
- **Isolated containers**: Each execution in fresh environment
- **Resource limits**: CPU/memory constraints via Testcontainers
- **No network access**: Containers run isolated by default  
- **Non-root execution**: Code runs as unprivileged user
- **Automatic timeout**: Built-in execution time limits
- **Automatic cleanup**: No leftover containers or resources

## Implementation Phases

### Phase 1: Core Functionality
1. Create `CodeExecutor` struct with Testcontainers integration
2. Implement language handlers (Python, Node.js, Go, Java)
3. Add resource limits and timeout handling
4. Create basic API interface

### Phase 2: Enhanced Features  
1. Add parallel execution support
2. Implement custom lifecycle hooks for security
3. Add input validation and sanitization
4. Create comprehensive error handling

### Phase 3: Advanced Options
1. Support for multi-file projects
2. Package/dependency installation
3. Real-time execution streaming
4. Metrics and monitoring

## File Structure
```
cmd/
  main.go                 # CLI interface
pkg/
  executor/
    executor.go           # Main executor using Testcontainers
    languages.go          # Language-specific configurations  
    security.go           # Security and resource management
  api/
    handlers.go           # HTTP API handlers
go.mod                   # Dependencies (testcontainers-go)
```

## Key Dependencies
- `github.com/testcontainers/testcontainers-go`
- `github.com/testcontainers/testcontainers-go/wait`

## Example Usage
```go
executor := NewCodeExecutor(30*time.Second)
result, err := executor.Execute(ctx, ExecutionRequest{
    Language: "python",
    Code: "print('Hello, World!')",
    Timeout: 10*time.Second,
})
```

## Research Summary

### Container Technologies Evaluated
1. **Docker**: Standard containerization with good isolation
2. **gVisor**: Enhanced security with syscall filtering (performance trade-off)
3. **Firecracker**: MicroVMs for maximum isolation (AWS Lambda-style)
4. **Kata Containers**: Lightweight VMs with container interface

### Security Considerations
- Container isolation provides baseline security
- Resource limits prevent resource exhaustion attacks
- Non-root execution reduces privilege escalation risks
- Network isolation prevents external communication
- Automatic cleanup prevents resource leaks

### Language Support Strategy
- Start with popular languages (Python, JavaScript, Go, Java)
- Use official Alpine-based images for minimal attack surface
- Implement language-specific execution patterns
- Support for package installation in future phases

This approach leverages Testcontainers' proven container management for a robust, secure code execution sandbox.