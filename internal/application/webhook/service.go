package webhook

// internal/http/middleware/versioning.go

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
)

type Service struct {
	repo       ports.WebhookRepository
	httpClient *http.Client
	queue      ports.EventPublisher
}

type WebhookPayload struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	ShopID    string          `json:"shop_id"`
	CreatedAt time.Time       `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

// WebhookDeliveryEvent is published when a webhook needs to be delivered
type WebhookDeliveryEvent struct {
	WebhookID uuid.UUID
	Payload   WebhookPayload
}

// TriggerWebhooks queues webhooks for a given topic
func (s *Service) TriggerWebhooks(ctx context.Context, shopID uuid.UUID, topic string, data any) error {
	webhooks, err := s.repo.ListByShopAndTopic(ctx, shopID, topic)
	if err != nil {
		return fmt.Errorf("list webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		return nil
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	payload := WebhookPayload{
		ID:        uuid.New().String(),
		Topic:     topic,
		ShopID:    shopID.String(),
		CreatedAt: time.Now(),
		Data:      dataJSON,
	}

	// Queue delivery for each webhook
	for _, wh := range webhooks {
		s.queue.Publish(ctx, WebhookDeliveryEvent{
			WebhookID: wh.ID,
			Payload:   payload,
		})
	}

	return nil
}

// DeliverWebhook actually sends the HTTP request
func (s *Service) DeliverWebhook(ctx context.Context, webhookID uuid.UUID, payload WebhookPayload) error {
	webhook, err := s.repo.GetByID(ctx, webhookID)
	if err != nil {
		return err
	}

	if webhook.Status != domain.WebhookStatusActive {
		return nil // Skip inactive webhooks
	}

	body, _ := json.Marshal(payload)

	// Sign the payload using the secret hash
	signature := s.signPayload(body, webhook.SecretHash)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.EndpointURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-ID", payload.ID)
	req.Header.Set("X-Webhook-Topic", payload.Topic)
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Timestamp", payload.CreatedAt.Format(time.RFC3339))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordFailure(ctx, webhook, err.Error())
		return fmt.Errorf("deliver webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.recordFailure(ctx, webhook, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return fmt.Errorf("webhook returned %d", resp.StatusCode)
	}

	// Reset failure count on success
	s.repo.ResetFailureCount(ctx, webhook.ID)

	return nil
}

func (s *Service) signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func (s *Service) recordFailure(ctx context.Context, webhook *domain.Webhook, reason string) {
	webhook.FailureCount++
	webhook.LastFailureAt = timePtr(time.Now())
	webhook.LastFailureReason = reason

	// Disable after 10 consecutive failures
	if webhook.FailureCount >= 10 {
		webhook.Status = domain.WebhookStatusDisabled
	}

	s.repo.Update(ctx, webhook)
}

func timePtr(t time.Time) *time.Time {
	return &t
}
