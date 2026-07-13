package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Webhook struct {
	ID             string   `json:"id"`
	URL            string   `json:"url"`
	Events         []string `json:"events"`
	Secret         string   `json:"secret,omitempty"`
	IsActive       bool     `json:"isActive"`
	LastDeliveryAt string   `json:"lastDeliveryAt,omitempty"`
	FailureCount   int      `json:"failureCount"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

type WebhookDelivery struct {
	ID           int    `json:"id"`
	WebhookID    string `json:"webhookId"`
	Event        string `json:"event"`
	Payload      string `json:"payload"`
	StatusCode   int    `json:"statusCode"`
	ResponseBody string `json:"responseBody"`
	Success      bool   `json:"success"`
	CreatedAt    string `json:"createdAt"`
}

func ListWebhooks(db *sql.DB) ([]Webhook, error) {
	rows, err := db.Query("SELECT id, url, events, secret, is_active, failure_count, created_at, updated_at FROM webhooks ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]Webhook, 0)
	for rows.Next() {
		var w Webhook
		var eventsJSON string
		if err := rows.Scan(&w.ID, &w.URL, &eventsJSON, &w.Secret, &w.IsActive, &w.FailureCount, &w.CreatedAt, &w.UpdatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(eventsJSON), &w.Events)
		results = append(results, w)
	}
	return results, nil
}

func SaveWebhook(db *sql.DB, w Webhook) error {
	eventsJSON, _ := json.Marshal(w.Events)
	_, err := db.Exec(
		"INSERT INTO webhooks (id, url, events, secret, is_active, failure_count, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(id) DO UPDATE SET url=excluded.url, events=excluded.events, secret=excluded.secret, is_active=excluded.is_active, updated_at=excluded.updated_at",
		w.ID, w.URL, string(eventsJSON), w.Secret, w.IsActive, w.FailureCount, time.Now().UTC().Format(time.RFC3339))
	return err
}

func DeleteWebhook(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM webhooks WHERE id = ?", id)
	return err
}

func RecordWebhookDelivery(db *sql.DB, d WebhookDelivery) error {
	_, err := db.Exec(
		"INSERT INTO webhook_deliveries (webhook_id, event, payload, status_code, response_body, success) VALUES (?, ?, ?, ?, ?, ?)",
		d.WebhookID, d.Event, d.Payload, d.StatusCode, d.ResponseBody, d.Success)
	return err
}

func IncrementWebhookFailure(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE webhooks SET failure_count = failure_count + 1, updated_at = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func ResetWebhookFailures(db *sql.DB, id string) error {
	_, err := db.Exec("UPDATE webhooks SET failure_count = 0, updated_at = ? WHERE id = ?",
		time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func GetWebhook(db *sql.DB, id string) (*Webhook, error) {
	var w Webhook
	var eventsJSON string
	err := db.QueryRow(`
		SELECT id, url, events, secret, is_active, failure_count, created_at, updated_at
		FROM webhooks WHERE id = ?
	`, id).Scan(&w.ID, &w.URL, &eventsJSON, &w.Secret, &w.IsActive, &w.FailureCount, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(eventsJSON), &w.Events)
	return &w, nil
}

func ListWebhookDeliveries(db *sql.DB, webhookID string, limit int) ([]WebhookDelivery, error) {
	rows, err := db.Query(`
		SELECT id, webhook_id, event, payload, status_code, response_body, success, created_at
		FROM webhook_deliveries WHERE webhook_id = ? ORDER BY created_at DESC, id DESC LIMIT ?
	`, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	deliveries := make([]WebhookDelivery, 0)
	for rows.Next() {
		var d WebhookDelivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.Event, &d.Payload, &d.StatusCode, &d.ResponseBody, &d.Success, &d.CreatedAt); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}
