package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestCreateHandler tests the /create endpoint
func TestCreateHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/create", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCreate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusFound)
	}

	location := rr.Header().Get("Location")
	if !strings.HasPrefix(location, "/interview/") {
		t.Errorf("handler returned wrong location header: got %v want /interview/...", location)
	}
}

// TestRootHandler tests the / endpoint
func TestRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRoot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// TestInterviewHandler tests the /interview/{id} endpoint
func TestInterviewHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/interview/12345", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleRoot)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// TestCompileHandler tests the /compile endpoint
func TestCompileHandler(t *testing.T) {
	// Case 1: Invalid Method
	req, err := http.NewRequest("GET", "/compile", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleCompile)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code for invalid method: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}

	// Case 2: Invalid JSON
	req, err = http.NewRequest("POST", "/compile", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid JSON: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Case 3: Valid Go Code (Simple)
	// Note: This requires 'go' to be in the path and working.
	code := `package main
	import "fmt"
	func main() {
		fmt.Println("Hello, World!")
	}`
	body, _ := json.Marshal(map[string]string{"code": code})
	req, err = http.NewRequest("POST", "/compile", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// If go is not installed or fails, this might fail.
	// But in this environment, go is installed.
	if status := rr.Code; status != http.StatusOK {
		// If it failed, check if it's because of missing go or compilation error
		// For integration test, we expect it to pass if environment is correct.
		// If it fails, we might want to log the body to see why.
		t.Logf("Compile failed with body: %s", rr.Body.String())
		// We won't fail the test hard if it's just environment issue, but for now let's assert success
		// assuming the environment is capable (which it is).
		t.Errorf("handler returned wrong status code for valid code: got %v want %v",
			status, http.StatusOK)
	}
	
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/wasm" {
		t.Errorf("handler returned wrong content type: got %v want application/wasm", contentType)
	}
}

// TestWebSocketHandler tests the /ws/{id} endpoint
func TestWebSocketHandler(t *testing.T) {
	// Setup HubManager
	hubManager := newHubManager()

	// Create a test server with the WebSocket handler logic
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from /ws/{id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "Invalid WebSocket URL", http.StatusBadRequest)
			return
		}
		id := parts[2]

		hub := hubManager.getOrCreateHub(id)
		serveWs(hub, w, r)
	}))
	defer server.Close()

	// Convert http URL to ws URL
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-session"

	// Connect to the WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer ws.Close()

	// Send a message
	message := []byte("hello world")
	if err := ws.WriteMessage(websocket.TextMessage, message); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Set read deadline to prevent hanging if no message is received
	ws.SetReadDeadline(time.Now().Add(time.Second * 5))

	// Read the message back (should be broadcasted to all clients, including sender)
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if !bytes.Equal(message, p) {
		t.Errorf("echo: got %s, want %s", p, message)
	}
}


// TestMultiClientJSExecution tests the full flow: create session, connect clients, sync code, execute JS
func TestMultiClientJSExecution(t *testing.T) {
	// 1. Setup Server
	hubManager := newHubManager()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/ws/") {
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) < 3 {
				http.Error(w, "Invalid WebSocket URL", http.StatusBadRequest)
				return
			}
			id := parts[2]
			hub := hubManager.getOrCreateHub(id)
			serveWs(hub, w, r)
			return
		}
		if r.URL.Path == "/compile" {
			handleCompile(w, r)
			return
		}
		if r.URL.Path == "/create" {
			handleCreate(w, r)
			return
		}
	}))
	defer server.Close()

	// 2. Create a new interview session
	resp, err := http.Get(server.URL + "/create")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		// handleCreate redirects, so client follows it. 
		// But wait, handleCreate redirects to /interview/{id}. 
		// Our test server doesn't handle /interview/{id} specifically in the mux above, 
		// but http.Get follows redirects.
		// Let's check the final URL to get the ID.
	}
	
	// Extract session ID from the URL
	// The URL will be something like http://127.0.0.1:port/interview/abcd
	parts := strings.Split(resp.Request.URL.Path, "/")
	sessionID := parts[len(parts)-1]
	if sessionID == "" {
		t.Fatalf("Failed to extract session ID from URL: %s", resp.Request.URL.Path)
	}

	// 3. Connect Client 1
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/" + sessionID
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Client 1 failed to connect: %v", err)
	}
	defer ws1.Close()

	// 4. Connect Client 2
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Client 2 failed to connect: %v", err)
	}
	defer ws2.Close()

	// 5. Client 1 sends valid JavaScript code
	jsCode := `console.log("Hello JS Integration Test")`
	msg := map[string]string{
		"type":    "code_update",
		"content": jsCode,
	}
	msgBytes, _ := json.Marshal(msg)
	if err := ws1.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		t.Fatalf("Client 1 failed to send message: %v", err)
	}

	// 6. Verify Client 2 receives the code update
	ws2.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, p, err := ws2.ReadMessage()
	if err != nil {
		t.Fatalf("Client 2 failed to read message: %v", err)
	}

	var receivedMsg map[string]string
	if err := json.Unmarshal(p, &receivedMsg); err != nil {
		t.Fatalf("Failed to unmarshal received message: %v", err)
	}

	if receivedMsg["content"] != jsCode {
		t.Errorf("Client 2 received wrong code: got %q, want %q", receivedMsg["content"], jsCode)
	}

	// 7. Compile/Execute the code
	// We use the same server URL for compile endpoint
	compileReq := map[string]string{
		"code":     jsCode,
		"language": "javascript",
	}
	compileBody, _ := json.Marshal(compileReq)
	resp, err = http.Post(server.URL+"/compile", "application/json", bytes.NewBuffer(compileBody))
	if err != nil {
		t.Fatalf("Failed to call /compile: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("/compile returned wrong status: got %v, want %v", resp.StatusCode, http.StatusOK)
	}

	// 8. Verify execution result
	// For JS, we expect text/plain output
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Wrong Content-Type: got %v, want text/plain", contentType)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	output := buf.String()

	if !strings.Contains(output, "Hello JS Integration Test") {
		t.Errorf("Execution output wrong: got %q, want it to contain 'Hello JS Integration Test'", output)
	}
}
