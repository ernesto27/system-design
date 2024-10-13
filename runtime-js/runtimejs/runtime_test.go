package runtimejs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestConsoleLog(t *testing.T) {
	content := `
		console.log("Hello, World!");
		console.log("The answer is", 42, true);
		console.log({ name: "John", age: 30 });
	`

	tmpfile := createTempFile(t, content)
	defer os.Remove(tmpfile)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	runtime, err := NewRuntimeJS(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create RuntimeJS: %v", err)
	}
	if runtime == nil {
		t.Fatal("RuntimeJS is nil")
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify the output
	expectedOutputs := []string{
		"Hello, World!",
		"The answer is 42 true",
		`{"name":"John","age":30}`,
	}

	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %s, actual %s", expected, output)
		}
	}
}

func TestSetTimeout(t *testing.T) {
	js := `
		let called = false;
		setTimeout(() => {
			called = true;
			console.log("Timeout executed");
		}, 100);
	`

	tmpfile := createTempFile(t, js)
	defer os.Remove(tmpfile)

	runtime, err := NewRuntimeJS(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create RuntimeJS: %v", err)
	}

	go runtime.RunEventLoop()

	time.Sleep(200 * time.Millisecond)

	called := runtime.vm.Get("called")

	if !called.ToBoolean() {
		t.Error("setTimeout callback was not executed")
	}

	close(runtime.done)
}

func TestSetInterval(t *testing.T) {
	content := `
		let count = 0;
		let intervalId = setInterval(() => {
			count++;
			if (count >= 3) {
				clearInterval(intervalId);
			}
		}, 100);
	`
	tempFilePath := createTempFile(t, content)
	defer os.Remove(tempFilePath)

	runtime, err := NewRuntimeJS(tempFilePath)
	if err != nil {
		t.Fatalf("Failed to create RuntimeJS: %v", err)
	}

	// Wait for the interval to execute multiple times
	time.Sleep(1000 * time.Millisecond)

	// Check the count
	count := runtime.vm.Get("count")

	countValue := count.ToInteger()

	var want int64 = 3

	if countValue != want {
		t.Errorf("setInterval did not execute as expected. Got count: %v, want: %d", countValue, want)
	}
}

func createTempFile(t *testing.T, content string) string {
	tempFile, err := os.CreateTemp("", "test.js")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tempFile.Name()
}
