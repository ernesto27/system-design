package main

import (
	"os"
	"testing"
	"time"
)

func TestEngine_Get(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	e := NewEngine(tmpfile.Name())
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key99", "value99")

	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Key exists",
			args: args{key: "key1"},
			want: "value1",
		},
		{
			name: "Key does not exist",
			args: args{key: "key4"},
			want: "",
		},
		{
			name: "Key 99 exists",
			args: args{key: "key99"},
			want: "value99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := e.Get(tt.args.key)

			if got != tt.want {
				t.Errorf("Expected %v, but got %v", tt.want, got)
			}
		})
	}
	defer e.Close()
}

func TestEngine_Compact(t *testing.T) {

	os.Remove("file_test.txt")
	v1 := "latestvalue1"
	v2 := "latestvalue2"
	e := NewEngine("file_test.txt")
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key1", v1)
	e.Set("key2", v2)
	e.Set("key3", "value3")

	go e.CompactFile()

	v := e.Get("key1")
	if v != v1 {
		t.Errorf("Expected %s, but got %s", v1, v)
	}

	v = e.Get("key2")
	if v != v2 {
		t.Errorf("Expected %s, but got %s", v2, v)
	}

	time.Sleep((Seconds * 2) * time.Second)

	expected := "key3:value3\nkey1:latestvalue1\nkey2:latestvalue2"

	// Check new file content
	// Fix this tests,  sometimes it fails, beacuse map order
	if e.GetFileContent() != expected {
		t.Errorf("Expected %s, but got %s", expected, e.GetFileContent())
	}

}

func TestEngine_Restore(t *testing.T) {
	os.Remove("file_test.txt")
	e := NewEngine("file_test.txt")

	e.Set("key1_restore", "value1")
	e.Set("key2_restore", "value2")

	e.Close()

	e = NewEngine("file_test.txt")
	e.Restore()
	k := e.Get("key1_restore")

	if k != "value1" {
		t.Errorf("Expected %s, but got %s", "value1", k)
	}
}
