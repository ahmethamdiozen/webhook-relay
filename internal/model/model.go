package model

import "time"

type Endpoint struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Target struct {
	ID         string    `json:"id"`
	EndpointID string    `json:"endpoint_id"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
}

type Event struct {
	ID         string    `json:"id"`
	EndpointID string    `json:"endpoint_id"`
	Headers    string    `json:"headers"`
	Body       string    `json:"body"`
	ReceivedAt time.Time `json:"received_at"`
}

type Delivery struct {
	ID         string    `json:"id"`
	EventID    string    `json:"event_id"`
	TargetID   string    `json:"target_id"`
	TargetURL  string    `json:"target_url"`
	Status     string    `json:"status"` // pending, success, failed
	Attempts   int       `json:"attempts"`
	LastError  string    `json:"last_error,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
