package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ResponseJson struct {
	Status  string `json:"key"`
	Message string `json:"value"`
}

func handlerSet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var rp RequestPayload

		err = json.Unmarshal(body, &rp)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
		}

		err = e.Set(rp.Key, rp.Value)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		responseJSON(w, ResponseJson{
			Status:  "success",
			Message: "Key value pair saved successfully.",
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func handlerGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		key := r.URL.Query().Get("key")
		value, err := e.Get(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusNotFound)
			return
		}

		responseJSON(w, RequestPayload{
			Key:   key,
			Value: value,
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func handlerDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		key := r.URL.Query().Get("key")
		err := e.Delete(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		responseJSON(w, ResponseJson{
			Status:  "success",
			Message: "Key deleted successfully.",
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func responseJSON(w http.ResponseWriter, data interface{}, status int) {
	d, err := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server errror"))
		return
	}

	w.WriteHeader(status)
	w.Write(d)
}

var e *Engine

func main() {
	e = NewEngine("db.txt", "delete.txt")
	defer e.Close()
	e.Restore()

	go e.CompactFile()
	go e.DeleteFromFile()

	http.HandleFunc("/set", handlerSet)
	http.HandleFunc("/get", handlerGet)
	http.HandleFunc("/delete", handlerDelete)

	address := ":8080"

	fmt.Printf("Server is listening on http://localhost%s\n", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("Error:", err)
	}

	// 	// remove file
	// 	//os.Remove("file.txt")
	// 	cache := NewEngine("db.txt")
	// 	defer cache.Close()
	// 	// cache.Set("key4", "value1")
	// 	// cache.Set("key6", "value2")
	// 	cache.Restore()

	// 	fmt.Println(cache.Get("key982873"))
	// 	return

	// 	// cache.Restore()
	// 	// fmt.Println(cache.Get("key7966"))
	// 	// return

	// 	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec a diam lectus. Sed sit amet ipsum mauris. "

	// 	for i := 0; i < 1000000; i++ {
	// 		cache.Set("key"+strconv.Itoa(i), lorem)
	// 	}

	// 	return

	// 	go cache.CompactFile()

	// 	exitSignal := make(chan struct{})
	// 	<-exitSignal

}
