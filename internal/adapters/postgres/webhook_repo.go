// internal/adapters/postgres/webhook_repo.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dfodeker/storeos/internal/adapters/postgres/db"
	"github.com/dfodeker/storeos/internal/domain"
	"github.com/dfodeker/storeos/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWebhookNotFound = errors.New("webhook not found")
)

type WebhookRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *WebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	row, err := r.queries.CreateWebhook(ctx, db.CreateWebhookParams{
		OrganizationID: webhook.OrganizationID,
		ShopID:         uuidPtrToPgtype(webhook.ShopID),
		Topic:          webhook.Topic,
		EndpointUrl:    webhook.EndpointURL,
		SecretHash:     webhook.SecretHash,
		Fields:         webhook.Fields,
		Status:         string(webhook.Status),
		ApiVersion:     webhook.APIVersion,
	})
	if err != nil {
		if isPgUniqueViolation(err) {
			return fmt.Errorf("webhook already exists for this topic and endpoint")
		}
		return fmt.Errorf("create webhook: %w", err)
	}

	webhook.ID = row.ID
	webhook.CreatedAt = row.CreatedAt
	webhook.UpdatedAt = row.UpdatedAt

	return nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	row, err := r.queries.GetWebhookByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWebhookNotFound
		}
		return nil, fmt.Errorf("get webhook: %w", err)
	}

	return rowToWebhook(row), nil
}

func (r *WebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	_, err := r.queries.UpdateWebhook(ctx, db.UpdateWebhookParams{
		ID:          webhook.ID,
		Topic:       textToPgtype(webhook.Topic),
		EndpointUrl: textToPgtype(webhook.EndpointURL),
		SecretHash:  textToPgtype(webhook.SecretHash),
		Fields:      webhook.Fields,
		Status:      textToPgtype(string(webhook.Status)),
		ApiVersion:  textToPgtype(webhook.APIVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWebhookNotFound
		}
		return fmt.Errorf("update webhook: %w", err)
	}

	return nil
}

func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteWebhook(ctx, id)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	return nil
}

func (r *WebhookRepository) ListByShop(ctx context.Context, shopID uuid.UUID, filter ports.WebhookFilter) ([]domain.Webhook, error) {
	limit := int32(filter.Limit)
	if limit == 0 {
		limit = 20
	}

	rows, err := r.queries.ListWebhooksByShop(ctx, db.ListWebhooksByShopParams{
		ShopID:  pgtype.UUID{Bytes: shopID, Valid: true},
		Column2: optionalStringValue(filter.Topic),
		Column3: webhookStatusToString(filter.Status),
		Limit:   limit,
		Offset:  int32(filter.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}

	webhooks := make([]domain.Webhook, len(rows))
	for i, row := range rows {
		webhooks[i] = *rowToWebhook(row)
	}

	return webhooks, nil
}

func (r *WebhookRepository) ListByShopAndTopic(ctx context.Context, shopID uuid.UUID, topic string) ([]domain.Webhook, error) {
	rows, err := r.queries.ListWebhooksByShopAndTopic(ctx, db.ListWebhooksByShopAndTopicParams{
		ShopID: pgtype.UUID{Bytes: shopID, Valid: true},
		Topic:  topic,
	})
	if err != nil {
		return nil, fmt.Errorf("list webhooks by topic: %w", err)
	}

	webhooks := make([]domain.Webhook, len(rows))
	for i, row := range rows {
		webhooks[i] = *rowToWebhook(row)
	}

	return webhooks, nil
}

func (r *WebhookRepository) ListByOrganization(ctx context.Context, orgID uuid.UUID, filter ports.WebhookFilter) ([]domain.Webhook, error) {
	limit := int32(filter.Limit)
	if limit == 0 {
		limit = 20
	}

	rows, err := r.queries.ListWebhooksByOrganization(ctx, db.ListWebhooksByOrganizationParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         int32(filter.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list webhooks by organization: %w", err)
	}

	webhooks := make([]domain.Webhook, len(rows))
	for i, row := range rows {
		webhooks[i] = *rowToWebhook(row)
	}

	return webhooks, nil
}

func (r *WebhookRepository) ListByOrganizationAndTopic(ctx context.Context, orgID uuid.UUID, topic string) ([]domain.Webhook, error) {
	rows, err := r.queries.ListWebhooksByOrganizationAndTopic(ctx, db.ListWebhooksByOrganizationAndTopicParams{
		OrganizationID: orgID,
		Topic:          topic,
	})
	if err != nil {
		return nil, fmt.Errorf("list webhooks by organization and topic: %w", err)
	}

	webhooks := make([]domain.Webhook, len(rows))
	for i, row := range rows {
		webhooks[i] = *rowToWebhook(row)
	}

	return webhooks, nil
}

func (r *WebhookRepository) IncrementFailureCount(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := r.queries.IncrementWebhookFailureCount(ctx, db.IncrementWebhookFailureCountParams{
		ID:                id,
		LastFailureReason: textToPgtype(reason),
	})
	if err != nil {
		return fmt.Errorf("increment failure count: %w", err)
	}
	return nil
}

func (r *WebhookRepository) ResetFailureCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.queries.ResetWebhookFailureCount(ctx, id)
	if err != nil {
		return fmt.Errorf("reset failure count: %w", err)
	}
	return nil
}

// Helper functions

func rowToWebhook(row db.Webhook) *domain.Webhook {
	webhook := &domain.Webhook{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Topic:          row.Topic,
		EndpointURL:    row.EndpointUrl,
		SecretHash:     row.SecretHash,
		Fields:         row.Fields,
		Status:         domain.WebhookStatus(row.Status),
		APIVersion:     row.ApiVersion,
		FailureCount:   int(row.FailureCount),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if row.ShopID.Valid {
		shopID := uuid.UUID(row.ShopID.Bytes)
		webhook.ShopID = &shopID
	}
	if row.LastFailureAt.Valid {
		webhook.LastFailureAt = &row.LastFailureAt.Time
	}
	if row.LastFailureReason.Valid {
		webhook.LastFailureReason = row.LastFailureReason.String
	}

	return webhook
}

// NOTE: use uuidPtrToPgtype from helpers.go for *uuid.UUID conversions

func optionalTextToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func webhookStatusToPgtype(s *domain.WebhookStatus) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: string(*s), Valid: true}
}

func optionalStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func webhookStatusToString(s *domain.WebhookStatus) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

// WebhookDeliveryRepository implementation

type WebhookDeliveryRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewWebhookDeliveryRepository(pool *pgxpool.Pool) *WebhookDeliveryRepository {
	return &WebhookDeliveryRepository{
		pool:    pool,
		queries: db.New(pool),
	}
}

func (r *WebhookDeliveryRepository) Create(ctx context.Context, delivery *domain.WebhookDelivery) error {
	headersJSON, err := json.Marshal(delivery.RequestHeaders)
	if err != nil {
		return fmt.Errorf("marshal request headers: %w", err)
	}

	row, err := r.queries.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
		WebhookID:      delivery.WebhookID,
		OrganizationID: delivery.OrganizationID,
		ShopID:         uuidPtrToPgtype(delivery.ShopID),
		Topic:          delivery.Topic,
		EndpointUrl:    delivery.EndpointURL,
		RequestHeaders: headersJSON,
		RequestBody:    delivery.RequestBody,
		Status:         string(delivery.Status),
	})
	if err != nil {
		return fmt.Errorf("create webhook delivery: %w", err)
	}

	delivery.ID = row.ID
	delivery.CreatedAt = row.CreatedAt

	return nil
}

func (r *WebhookDeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WebhookDelivery, error) {
	row, err := r.queries.GetWebhookDeliveryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("webhook delivery not found")
		}
		return nil, fmt.Errorf("get webhook delivery: %w", err)
	}

	return rowToWebhookDelivery(row), nil
}

func (r *WebhookDeliveryRepository) Update(ctx context.Context, delivery *domain.WebhookDelivery) error {
	responseHeadersJSON, _ := json.Marshal(delivery.ResponseHeaders)

	_, err := r.queries.UpdateWebhookDelivery(ctx, db.UpdateWebhookDeliveryParams{
		ID:              delivery.ID,
		ResponseStatus:  pgtype.Int4{Int32: int32(delivery.ResponseStatus), Valid: delivery.ResponseStatus != 0},
		ResponseHeaders: responseHeadersJSON,
		ResponseBody:    delivery.ResponseBody,
		Status:          textToPgtype(string(delivery.Status)),
		Attempts:        pgtype.Int4{Int32: int32(delivery.Attempts), Valid: true},
		NextRetryAt:     timeToPgtype(delivery.NextRetryAt),
		ErrorMessage:    textToPgtype(delivery.ErrorMessage),
		DurationMs:      pgtype.Int4{Int32: int32(delivery.DurationMs), Valid: delivery.DurationMs != 0},
		DeliveredAt:     timeToPgtype(delivery.DeliveredAt),
	})
	if err != nil {
		return fmt.Errorf("update webhook delivery: %w", err)
	}

	return nil
}

func (r *WebhookDeliveryRepository) ListByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]domain.WebhookDelivery, error) {
	rows, err := r.queries.ListWebhookDeliveriesByWebhook(ctx, db.ListWebhookDeliveriesByWebhookParams{
		WebhookID: webhookID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list webhook deliveries: %w", err)
	}

	deliveries := make([]domain.WebhookDelivery, len(rows))
	for i, row := range rows {
		deliveries[i] = *rowToWebhookDelivery(row)
	}

	return deliveries, nil
}

func (r *WebhookDeliveryRepository) ListPendingRetries(ctx context.Context, limit int) ([]domain.WebhookDelivery, error) {
	rows, err := r.queries.ListPendingWebhookDeliveries(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("list pending deliveries: %w", err)
	}

	deliveries := make([]domain.WebhookDelivery, len(rows))
	for i, row := range rows {
		deliveries[i] = *rowToWebhookDelivery(row)
	}

	return deliveries, nil
}

func (r *WebhookDeliveryRepository) MarkAsSuccess(ctx context.Context, id uuid.UUID, responseStatus int, responseBody []byte, durationMs int) error {
	_, err := r.queries.MarkWebhookDeliverySuccess(ctx, db.MarkWebhookDeliverySuccessParams{
		ID:             id,
		ResponseStatus: pgtype.Int4{Int32: int32(responseStatus), Valid: true},
		ResponseBody:   responseBody,
		DurationMs:     pgtype.Int4{Int32: int32(durationMs), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("mark delivery success: %w", err)
	}
	return nil
}

func (r *WebhookDeliveryRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string, nextRetryAt *time.Time) error {
	_, err := r.queries.MarkWebhookDeliveryFailed(ctx, db.MarkWebhookDeliveryFailedParams{
		ID:           id,
		ErrorMessage: textToPgtype(errorMessage),
		NextRetryAt:  timeToPgtype(nextRetryAt),
	})
	if err != nil {
		return fmt.Errorf("mark delivery failed: %w", err)
	}
	return nil
}

// Helper functions for webhook delivery

func rowToWebhookDelivery(row db.WebhookDelivery) *domain.WebhookDelivery {
	delivery := &domain.WebhookDelivery{
		ID:             row.ID,
		WebhookID:      row.WebhookID,
		OrganizationID: row.OrganizationID,
		Topic:          row.Topic,
		EndpointURL:    row.EndpointUrl,
		RequestBody:    row.RequestBody,
		Status:         domain.WebhookDeliveryStatus(row.Status),
		Attempts:       int(row.Attempts),
		CreatedAt:      row.CreatedAt,
	}

	if row.ShopID.Valid {
		shopID := uuid.UUID(row.ShopID.Bytes)
		delivery.ShopID = &shopID
	}

	// Parse request headers
	if row.RequestHeaders != nil {
		json.Unmarshal(row.RequestHeaders, &delivery.RequestHeaders)
	}

	// Parse response headers
	if row.ResponseHeaders != nil {
		json.Unmarshal(row.ResponseHeaders, &delivery.ResponseHeaders)
	}

	if row.ResponseStatus.Valid {
		delivery.ResponseStatus = int(row.ResponseStatus.Int32)
	}
	if row.ResponseBody != nil {
		delivery.ResponseBody = row.ResponseBody
	}
	if row.NextRetryAt.Valid {
		delivery.NextRetryAt = &row.NextRetryAt.Time
	}
	if row.ErrorMessage.Valid {
		delivery.ErrorMessage = row.ErrorMessage.String
	}
	if row.DurationMs.Valid {
		delivery.DurationMs = int(row.DurationMs.Int32)
	}
	if row.DeliveredAt.Valid {
		delivery.DeliveredAt = &row.DeliveredAt.Time
	}

	return delivery
}

func timeToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
