package main


import (
	"crypto/rand"
	"encoding/json"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// HubManager manages active hubs for different interview sessions
type HubManager struct {
	hubs map[string]*Hub
	mu   sync.RWMutex
}

func newHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*Hub),
	}
}

func (hm *HubManager) getOrCreateHub(id string) *Hub {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hub, ok := hm.hubs[id]; ok {
		return hub
	}

	hub := newHub()
	go hub.run()
	hm.hubs[id] = hub
	return hub
}

func generateID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "default"
	}
	return hex.EncodeToString(bytes)
}

func main() {
	hubManager := newHubManager()

	// Serve static files from the "static" directory
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Root handler - serve index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		// Handle /interview/{id}
		if strings.HasPrefix(r.URL.Path, "/interview/") {
			http.ServeFile(w, r, "./static/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// Create new interview session
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		id := generateID()
		http.Redirect(w, r, "/interview/"+id, http.StatusFound)
	})

	// WebSocket handler
	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from /ws/{id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "Invalid WebSocket URL", http.StatusBadRequest)
			return
		}
		id := parts[2]
		
		hub := hubManager.getOrCreateHub(id)
		serveWs(hub, w, r)
	})

	http.HandleFunc("/compile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create temp dir
		tmpDir, err := os.MkdirTemp("", "wasm-build-*")
		if err != nil {
			http.Error(w, "Failed to create temp dir", http.StatusInternalServerError)
			log.Printf("Error creating temp dir: %v", err)
			return
		}
		defer os.RemoveAll(tmpDir)

		// Write code to main.go
		if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(req.Code), 0644); err != nil {
			http.Error(w, "Failed to write code", http.StatusInternalServerError)
			log.Printf("Error writing code: %v", err)
			return
		}

		// Run go build
		cmd := exec.Command("go", "build", "-o", "main.wasm", "main.go")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Compilation error
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(output)
			return
		}

		// Read wasm file
		wasmBytes, err := os.ReadFile(filepath.Join(tmpDir, "main.wasm"))
		if err != nil {
			http.Error(w, "Failed to read wasm", http.StatusInternalServerError)
			log.Printf("Error reading wasm: %v", err)
			return
		}

		// Send wasm
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(wasmBytes)
	})

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
