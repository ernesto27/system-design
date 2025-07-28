package codeexecutor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestCaseData represents a single test case
type TestCaseData struct {
	Name           string                 `json:"name"`
	Parameters     map[string]interface{} `json:"parameters"`
	ExpectedOutput interface{}            `json:"expected_output"`
}

// TestResult represents the result of running a test case
type TestResult struct {
	Name     string      `json:"name"`
	Passed   bool        `json:"passed"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Error    string      `json:"error,omitempty"`
}

// TestResults represents all test results
type TestResults struct {
	Passed      int          `json:"passed"`
	Total       int          `json:"total"`
	TestResults []TestResult `json:"test_results"`
}

// CodeExecutor interface for executing code in different languages
type CodeExecutor interface {
	Execute(code string) (string, error)
	Language() string
}

// BaseExecutor contains common container execution logic
type BaseExecutor struct {
	image     string
	extension string
	command   []string
	language  string
}

// Execute runs code using the configured container settings
func (e *BaseExecutor) Execute(code string) (string, error) {
	ctx := context.Background()

	// Write to temp file with language-specific extension
	tmpFile, err := os.CreateTemp("", "script-*"+e.extension)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(code)
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		return "", err
	}

	// Build command with script path
	cmd := append(e.command, "/usr/src/app/script"+e.extension)

	req := testcontainers.ContainerRequest{
		Image:      e.image,
		Cmd:        cmd,
		WaitingFor: wait.ForExit(),
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      absPath,
				ContainerFilePath: "/usr/src/app/script" + e.extension,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", err
	}
	defer container.Terminate(ctx)

	// Wait for container to exit (with timeout)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err = req.WaitingFor.WaitUntilReady(ctxWithTimeout, container)
	if err != nil {
		return "", fmt.Errorf("container execution timeout or failed: %w", err)
	}

	// Read logs
	logs, err := container.Logs(ctx)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(logs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Language returns the language name
func (e *BaseExecutor) Language() string {
	return e.language
}

// JavaScript executor
type JavaScriptExecutor struct {
	BaseExecutor
}

// Python executor
type PythonExecutor struct {
	BaseExecutor
}

// Go executor
type GoExecutor struct {
	BaseExecutor
}

// NewExecutor creates a new executor for the specified language
func NewExecutor(language string) (CodeExecutor, error) {
	switch language {
	case "javascript", "js":
		return &JavaScriptExecutor{
			BaseExecutor{
				image:     "node:20",
				extension: ".js",
				command:   []string{"node"},
				language:  "javascript",
			},
		}, nil
	case "python", "py":
		return &PythonExecutor{
			BaseExecutor{
				image:     "python:3.11",
				extension: ".py",
				command:   []string{"python"},
				language:  "python",
			},
		}, nil
	case "go", "golang":
		return &GoExecutor{
			BaseExecutor{
				image:     "golang:1.21",
				extension: ".go",
				command:   []string{"go", "run"},
				language:  "go",
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// Execute function for backward compatibility
func Execute(code string) (string, error) {
	executor, err := NewExecutor("javascript")
	if err != nil {
		return "", err
	}
	return executor.Execute(code)
}
