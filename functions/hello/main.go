package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// handle answers POST / — it reads the event body and echoes a greeting.
// This is the whole function contract: event in (request body), result out
// (response body).
func handle(w http.ResponseWriter, r *http.Request) {
	event, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read event", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "hello, %s", event)
}

func main() {
	http.HandleFunc("/", handle)

	log.Println("function listening on :8080")
	// ListenAndServe blocks; each request runs handle in its own goroutine.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
