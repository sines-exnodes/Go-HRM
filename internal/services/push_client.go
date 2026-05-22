package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/google"

	"github.com/exnodes/hrm-api/internal/config"
)

// PushMessage is the provider-neutral push payload.
type PushMessage struct {
	Token string // FCM device token
	Title string
	Body  string
	Data  map[string]any
}

// PushClient is the provider-neutral interface. The default impl is
// FCM HTTP v1; tests can swap a fake.
type PushClient interface {
	// Send delivers a single message. Returns nil on success; non-nil
	// error means the caller should record it (the notification
	// service aggregates errors across multiple tokens).
	Send(ctx context.Context, msg PushMessage) error
	// IsConfigured reports whether the client has real backing.
	// no-op clients return false.
	IsConfigured() bool
}

// ---- No-op client (default when FIREBASE_CREDENTIALS_PATH is empty) ----

type noopPushClient struct{}

// Send logs the message and returns nil — keeps the service flow
// consistent in dev environments without Firebase.
func (n *noopPushClient) Send(ctx context.Context, msg PushMessage) error {
	log.Printf("push: skipped (no-op) token=%s title=%q body=%q", maskToken(msg.Token), msg.Title, msg.Body)
	return nil
}

func (n *noopPushClient) IsConfigured() bool { return false }

func maskToken(t string) string {
	if len(t) < 12 {
		return "***"
	}
	return t[:6] + "…" + t[len(t)-4:]
}

// ---- FCM HTTP v1 client ----

const fcmEndpointFormat = "https://fcm.googleapis.com/v1/projects/%s/messages:send"

type fcmPushClient struct {
	projectID string
	// http.Client with a goroutine-safe token source. oauth2's
	// TokenSource caches and refreshes automatically.
	httpClient *http.Client
}

// NewPushClient constructs the right implementation based on the
// configuration. When FIREBASE_CREDENTIALS_PATH is empty OR the file
// can't be read, returns the no-op client and logs a warning (does NOT
// fail the boot — REVISION NOTES #11).
func NewPushClient(cfg *config.Config) PushClient {
	credPath := strings.TrimSpace(cfg.FirebaseCredentialsPath)
	projectID := strings.TrimSpace(cfg.FirebaseProjectID)
	if credPath == "" || projectID == "" {
		log.Printf("push: FCM disabled — FIREBASE_CREDENTIALS_PATH or FIREBASE_PROJECT_ID is empty")
		return &noopPushClient{}
	}

	credBytes, err := os.ReadFile(credPath)
	if err != nil {
		log.Printf("push: FCM disabled — cannot read credentials at %s: %v", credPath, err)
		return &noopPushClient{}
	}

	jwtCfg, err := google.JWTConfigFromJSON(credBytes, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		log.Printf("push: FCM disabled — credentials parse failed: %v", err)
		return &noopPushClient{}
	}

	httpClient := jwtCfg.Client(context.Background())
	httpClient.Timeout = 10 * time.Second
	return &fcmPushClient{projectID: projectID, httpClient: httpClient}
}

// Send posts to FCM HTTP v1.
func (c *fcmPushClient) Send(ctx context.Context, msg PushMessage) error {
	if strings.TrimSpace(msg.Token) == "" {
		return errors.New("fcm: empty device token")
	}

	// FCM accepts only string values in `data` — coerce.
	data := make(map[string]string, len(msg.Data))
	for k, v := range msg.Data {
		switch val := v.(type) {
		case string:
			data[k] = val
		case fmt.Stringer:
			data[k] = val.String()
		default:
			data[k] = fmt.Sprintf("%v", val)
		}
	}

	payload := map[string]any{
		"message": map[string]any{
			"token": msg.Token,
			"notification": map[string]any{
				"title": msg.Title,
				"body":  msg.Body,
			},
			"data": data,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fcm: marshal: %w", err)
	}

	endpoint := fmt.Sprintf(fcmEndpointFormat, c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fcm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fcm: http: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fcm: status %d: %s", resp.StatusCode, string(rb))
	}
	return nil
}

func (c *fcmPushClient) IsConfigured() bool { return true }
