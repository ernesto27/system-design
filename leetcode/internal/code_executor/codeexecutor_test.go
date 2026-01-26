package codeexecutor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute_SimpleJavaScript(t *testing.T) {
	code := `console.log("Hello, World!");`
	
	output, err := Execute(code)
	assert.NoError(t, err)
	assert.Contains(t, output, "Hello, World!")
}

func TestNewExecutor_JavaScript(t *testing.T) {
	executor, err := NewExecutor("javascript")
	assert.NoError(t, err)
	assert.Equal(t, "javascript", executor.Language())
	
	code := `console.log("Hello from JS!");`
	output, err := executor.Execute(code)
	assert.NoError(t, err)
	assert.Contains(t, output, "Hello from JS!")
}

func TestNewExecutor_Python(t *testing.T) {
	executor, err := NewExecutor("python")
	assert.NoError(t, err)
	assert.Equal(t, "python", executor.Language())
	
	code := `print("Hello from Python!")`
	output, err := executor.Execute(code)
	assert.NoError(t, err)
	assert.Contains(t, output, "Hello from Python!")
}

func TestNewExecutor_Go(t *testing.T) {
	executor, err := NewExecutor("go")
	assert.NoError(t, err)
	assert.Equal(t, "go", executor.Language())
	
	code := `package main
import "fmt"
func main() {
	fmt.Println("Hello from Go!")
}`
	output, err := executor.Execute(code)
	assert.NoError(t, err)
	assert.Contains(t, output, "Hello from Go!")
}

func TestNewExecutor_UnsupportedLanguage(t *testing.T) {
	executor, err := NewExecutor("rust")
	assert.Error(t, err)
	assert.Nil(t, executor)
	assert.Contains(t, err.Error(), "unsupported language: rust")
}

func TestNewExecutor_LanguageAliases(t *testing.T) {
	// Test JavaScript aliases
	jsExecutor, err := NewExecutor("js")
	assert.NoError(t, err)
	assert.Equal(t, "javascript", jsExecutor.Language())
	
	// Test Python aliases
	pyExecutor, err := NewExecutor("py")
	assert.NoError(t, err)
	assert.Equal(t, "python", pyExecutor.Language())
	
	// Test Go aliases
	goExecutor, err := NewExecutor("golang")
	assert.NoError(t, err)
	assert.Equal(t, "go", goExecutor.Language())
}