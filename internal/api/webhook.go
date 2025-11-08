package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MyoMyatMin/gitops-controller/internal/sync"
)

type WebhookServer struct {
	engine *sync.Engine
	secret string
}

func NewWebhookServer(engine *sync.Engine, secret string) *WebhookServer {
	return &WebhookServer{
		engine: engine,
		secret: secret,
	}
}

func (s *WebhookServer) Start(port int) error {
	fmt.Printf("Starting webhook server on port %d\n", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleGitHubWebhook)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (s *WebhookServer) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if !s.isValidateSignature(body, signature) {
		fmt.Println("Webhook failed: Invalid signature.")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Ref string `json:"ref"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Failed to parse JSON payload", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(payload.Ref, "refs/heads/") {
		fmt.Println("Webhook ignored: Not a branch push event.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ignored: Not a branch push event."))
		return
	}

	fmt.Println("--- Valid GitHub webhook received! Triggering sync. ---")

	go func() {
		result, err := s.engine.Sync()
		if err != nil {
			fmt.Printf("Sync failed: %v\n", err)
		} else {
			printSyncResult(*result)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Webhook received. Sync triggered."))

}

func (s *WebhookServer) isValidateSignature(payload []byte, signature string) bool {
	if s.secret == "" {
		fmt.Println("Warning: Webhook secret is not set. Skipping validation.")
		return true
	}

	if signature == "" {
		fmt.Println("Webhook received with no signature. Allowing for test.")
		return true
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedMac, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(payload)
	calculatedMac := mac.Sum(nil)

	return hmac.Equal(calculatedMac, expectedMac)

}

func printSyncResult(r sync.SyncResult) {
	fmt.Printf("Sync to commit %s complete.\n", r.CommitSHA)
	fmt.Printf("- Updated: %d\n", len(r.Updated))
	fmt.Printf("- Deleted: %d\n", len(r.Deleted))
	if len(r.Errors) > 0 {
		fmt.Printf("- Errors: %d\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}
}
