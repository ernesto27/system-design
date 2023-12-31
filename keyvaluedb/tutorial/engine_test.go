package main

import (
	"os"
	"testing"
	"time"
)

func Test_SetGetKeyValue(t *testing.T) {
	e, _ := NewEngine()
	e.Set("test", "data")
	e.Set("foo", "bar")
	value, err := e.Get("foo")
	if err != nil {
		t.Error(err)
	}
	if value != "bar" {
		t.Error("value should be bar")
	}

	_, err = e.Get("notfound")
	if err == nil {
		t.Error("should return error")
	}
}

func TestEngine_Compact(t *testing.T) {
	os.Remove("data.txt")
	v1 := "latestvalue1"
	v2 := "latestvalue2"
	e, _ := NewEngine()
	e.Set("key1", "value1")
	e.Set("key2", "value2")
	e.Set("key1", v1)
	e.Set("key2", v2)
	e.Set("key3", "value3")

	go e.CompactFile()

	time.Sleep((Seconds + 3) * time.Second)

	if len(e.GetFileContent(e.file)) != 32 {
		t.Errorf("Expected %d, but got %d", 3, len(e.GetFileContent(e.file)))
	}

}
