package main


import (
	"crypto/rand"
	"encoding/json"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
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

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Parse JSON body
		var req struct {
			Code     string `json:"code"`
			Language string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Simulate execution
		time.Sleep(1 * time.Second) // Simulate build/run time

		var output string
		switch req.Language {
		case "go":
			output = "Building...\nRunning...\n\nProgram exited successfully.\nOutput:\nHello, World! (from Go)"
		case "javascript":
			output = "Running...\n\nOutput:\nHello, World! (from JavaScript)"
		case "python":
			output = "Running...\n\nOutput:\nHello, World! (from Python)"
		default:
			output = "Unknown language"
		}

		w.Write([]byte(output))
	})

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
