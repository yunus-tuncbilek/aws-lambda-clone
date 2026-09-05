package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// registry maps a function name to its container image.
// Phase 2: hardcoded. Later this becomes a real deploy API.
var registry = map[string]string{
	"hello": "docker.io/library/hello:latest",
}

// runner executes functions as containers (set up in main).
var runner *executor

// invoke handles POST /invoke/{name}: look up the image, run it in a fresh
// container, forward the event, return the response.
func invoke(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/invoke/")
	imageRef, ok := registry[name]
	if !ok {
		http.Error(w, "unknown function: "+name, http.StatusNotFound)
		return
	}

	event, _ := io.ReadAll(r.Body)

	body, status, err := runner.run(imageRef, event)
	if err != nil {
		log.Printf("invoke %s failed: %v", name, err)
		http.Error(w, "invocation failed", http.StatusBadGateway)
		return
	}

	w.WriteHeader(status)
	w.Write(body)
}

// waitReady polls addr until a TCP connection succeeds or timeout elapses.
func waitReady(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return &net.OpError{Op: "dial", Net: "tcp"}
}

func main() {
	var err error
	runner, err = newExecutor()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/invoke/", invoke)

	log.Println("control plane listening on :9000")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
