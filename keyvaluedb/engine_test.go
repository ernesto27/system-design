package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

const fileData = "file_test.txt"
const fileDelete = "delete_test.txt"

func TestEngine_Get(t *testing.T) {
	os.Remove(fileData)

	type args struct {
		key string
	}

	type user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	u := user{ID: 1, Name: "ernesto"}
	j, _ := json.Marshal(u)

	e, _ := NewEngine(fileData, fileDelete)
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key99", "value99")
	e.Set("json", string(j))

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
		{
			name: "set json",
			args: args{key: "json"},
			want: `{"id":1,"name":"ernesto"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, _ := e.Get(tt.args.key)

			if got != tt.want {
				t.Errorf("Expected %v, but got %v", tt.want, got)
			}
		})
	}
	defer e.Close()
}

func TestEngine_SetError(t *testing.T) {
	e, _ := NewEngine(fileData, fileDelete)
	err := e.Set("ke y1", "value1")

	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

}

func TestEngine_Compact(t *testing.T) {
	os.Remove(fileData)
	v1 := "latestvalue1"
	v2 := "latestvalue2"
	e, _ := NewEngine(fileData, fileDelete)
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key1", v1)
	e.Set("key2", v2)
	e.Set("key3", "value3")

	go e.CompactFile()

	v, _ := e.Get("key1")
	if v != v1 {
		t.Errorf("Expected %s, but got %s", v1, v)
	}

	v, _ = e.Get("key2")
	if v != v2 {
		t.Errorf("Expected %s, but got %s", v2, v)
	}

	time.Sleep((Seconds + 3) * time.Second)

	if len(e.GetFileContent(e.file)) != 3 {
		t.Errorf("Expected %d, but got %d", 3, len(e.GetFileContent(e.file)))
	}

}

func TestEngine_Restore(t *testing.T) {
	os.Remove(fileData)
	e, _ := NewEngine(fileData, fileDelete)

	e.Set("key1_restore", "value1")
	e.Set("key2_restore", "value2")

	e.Close()

	e, _ = NewEngine(fileData, fileDelete)
	e.Restore()
	k, _ := e.Get("key1_restore")

	if k != "value1" {
		t.Errorf("Expected %s, but got %s", "value1", k)
	}
}

func TestEngine_DeleteKey(t *testing.T) {
	os.Remove(fileData)
	os.Remove(fileDelete)
	e, _ := NewEngine(fileData, fileDelete)

	e.Set("key1_delete", "value1")
	e.Set("key2_delete", "value2")

	err := e.Delete("key1_delete")
	if err != nil {
		panic(err)
	}

	k, _ := e.Get("key1_delete")

	if k != "" {
		t.Errorf("Expected %s, but got %s", "", k)
	}

	if len(e.GetFileContent(e.fileDelete)) != 1 {
		t.Errorf("Expected %d, but got %d", 1, len(e.GetFileContent(e.file)))
	}
}

func TestEngine_DeleteFromFile(t *testing.T) {
	os.Remove(fileData)
	os.Remove(fileDelete)
	e, _ := NewEngine(fileData, fileDelete)

	e.Set("key1_delete", "value1")
	e.Set("key2_delete", "value2")
	e.Set("key3_delete", "value3")

	err := e.Delete("key2_delete")
	if err != nil {
		panic(err)
	}

	if len(e.GetFileContent(e.fileDelete)) != 1 {
		t.Errorf("Expected %d, but got %d", 1, len(e.GetFileContent(e.file)))
	}

	go e.DeleteFromFile()
	time.Sleep((Seconds + 3) * time.Second)

	if len(e.GetFileContent(e.fileDelete)) != 0 {
		t.Errorf("Expected %d, but got %d", 0, len(e.GetFileContent(e.fileDelete)))
	}
}

func TestEngine_DeleteKeyFromFile(t *testing.T) {
	os.Remove(fileData)
	os.Remove(fileDelete)
	e, _ := NewEngine(fileData, fileDelete)

	e.Set("key1_delete", "value1")
	e.Set("key2_delete", "value2")
	e.Set("key3_delete", "value3")

	e.deleteKeyFromFile([]string{"key2_delete", "key3_delete"})

	if len(e.GetFileContent(e.file)) != 1 {
		t.Errorf("Expected %d, but got %d", 1, len(e.GetFileContent(e.file)))
	}

}
