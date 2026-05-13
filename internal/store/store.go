package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ahmethamdiozen/webhook-relay/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) CreateEndpoint(ctx context.Context, name string) (*model.Endpoint, error) {
	var e model.Endpoint
	err := s.db.QueryRow(ctx,
		`INSERT INTO endpoints (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&e.ID, &e.Name, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create endpoint: %w", err)
	}
	return &e, nil
}

func (s *Store) ListEndpoints(ctx context.Context) ([]model.Endpoint, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, created_at FROM endpoints ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []model.Endpoint
	for rows.Next() {
		var e model.Endpoint
		if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, e)
	}
	return endpoints, nil
}

func (s *Store) CreateTarget(ctx context.Context, endpointID, url string) (*model.Target, error) {
	var t model.Target
	err := s.db.QueryRow(ctx,
		`INSERT INTO targets (endpoint_id, url) VALUES ($1, $2) RETURNING id, endpoint_id, url, created_at`,
		endpointID, url,
	).Scan(&t.ID, &t.EndpointID, &t.URL, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create target: %w", err)
	}
	return &t, nil
}

func (s *Store) GetTargetsByEndpoint(ctx context.Context, endpointID string) ([]model.Target, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, endpoint_id, url, created_at FROM targets WHERE endpoint_id = $1`,
		endpointID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []model.Target
	for rows.Next() {
		var t model.Target
		if err := rows.Scan(&t.ID, &t.EndpointID, &t.URL, &t.CreatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (s *Store) SaveEvent(ctx context.Context, endpointID string, r *http.Request, body []byte) (*model.Event, error) {
	headersJSON, _ := json.Marshal(r.Header)

	var e model.Event
	err := s.db.QueryRow(ctx,
		`INSERT INTO events (endpoint_id, headers, body) VALUES ($1, $2, $3) RETURNING id, endpoint_id, headers, body, received_at`,
		endpointID, string(headersJSON), string(body),
	).Scan(&e.ID, &e.EndpointID, &e.Headers, &e.Body, &e.ReceivedAt)
	if err != nil {
		return nil, fmt.Errorf("save event: %w", err)
	}
	return &e, nil
}

func (s *Store) CreateDelivery(ctx context.Context, eventID, targetID, targetURL string) (*model.Delivery, error) {
	var d model.Delivery
	err := s.db.QueryRow(ctx,
		`INSERT INTO deliveries (event_id, target_id, target_url) VALUES ($1, $2, $3) RETURNING id, event_id, target_id, target_url, status, attempts, created_at`,
		eventID, targetID, targetURL,
	).Scan(&d.ID, &d.EventID, &d.TargetID, &d.TargetURL, &d.Status, &d.Attempts, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create delivery: %w", err)
	}
	return &d, nil
}

func (s *Store) UpdateDelivery(ctx context.Context, id, status, lastError string, attempts int) error {
	_, err := s.db.Exec(ctx,
		`UPDATE deliveries SET status=$1, last_error=$2, attempts=$3, delivered_at=CASE WHEN $1='success' THEN now() ELSE NULL END WHERE id=$4`,
		status, lastError, attempts, id,
	)
	return err
}

func (s *Store) ListEvents(ctx context.Context) ([]model.Event, error) {
	rows, err := s.db.Query(ctx, `SELECT id, endpoint_id, headers, body, received_at FROM events ORDER BY received_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.EndpointID, &e.Headers, &e.Body, &e.ReceivedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}
