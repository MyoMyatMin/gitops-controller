package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/MyoMyatMin/gitops-controller/internal/sync"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
	log.Infof("Starting webhook server on port %d...", port)
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", s.handleGitHubWebhook)

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", s.handleHealth)

	mux.HandleFunc("/ready", s.handleHealth)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (s *WebhookServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
func (s *WebhookServer) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		log.Warnf("Invalid webhook method: %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Error reading webhook body: %v", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if !s.isValidSignature(body, signature) {
		log.Warn("Webhook failed: Invalid signature")
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Errorf("Error parsing webhook JSON payload: %v", err)
		http.Error(w, "Error parsing JSON payload", http.StatusBadRequest)
		return
	}

	ref := payload.Ref
	logFields := logrus.Fields{"ref": ref}

	if !strings.HasPrefix(ref, "refs/heads/") {
		log.WithFields(logFields).Info("Webhook ignored: Not a branch push event.")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ignored: Not a branch push event."))
		return
	}

	log.WithFields(logFields).Info("--- Valid GitHub webhook received! Triggering sync. ---")

	go func() {
		result, err := s.engine.Sync()
		if err != nil {
			log.Errorf("Webhook-triggered sync failed: %v", err)
		} else {

			log.WithFields(logrus.Fields{
				"commit":  result.CommitSHA,
				"updated": len(result.Updated),
				"deleted": len(result.Deleted),
				"errors":  len(result.Errors),
			}).Info("Webhook-triggered sync successful.")
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Accepted: Sync triggered."))
}

func (s *WebhookServer) isValidSignature(body []byte, signature string) bool {
	if s.secret == "" {
		log.Warn("Webhook secret is not set. Skipping validation.")
		return true
	}
	if signature == "" {
		log.Info("Webhook received with no signature. Allowing for test.")
		return true
	}
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	expectedMAC, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write(body)
	actualMAC := mac.Sum(nil)

	return hmac.Equal(actualMAC, expectedMAC)
}
