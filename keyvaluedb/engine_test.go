package main

import (
	"bytes"
	"os"
	"testing"
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
		want []byte
	}{
		{
			name: "Key exists",
			args: args{key: "key1"},
			want: []byte("value1"),
		},
		{
			name: "Key does not exist",
			args: args{key: "key4"},
			want: []byte{},
		},
		{
			name: "Key 99 exists",
			args: args{key: "key99"},
			want: []byte("value99"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := e.Get(tt.args.key)

			if !bytes.Equal(got, tt.want) {
				t.Errorf("Expected %v, but got %v", tt.want, got)
			}
		})
	}
	defer e.Close()
}
