package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ahmethamdiozen/webhook-relay/internal/relay"
	"github.com/ahmethamdiozen/webhook-relay/internal/store"
)

type Handler struct {
	store *store.Store
	relay *relay.Relay
}

func New(s *store.Store, r *relay.Relay) *Handler {
	return &Handler{store: s, relay: r}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/in/", h.ReceiveWebhook)
	mux.HandleFunc("/endpoints", h.Endpoints)
	mux.HandleFunc("/endpoints/", h.EndpointTargets)
	mux.HandleFunc("/events", h.ListEvents)
}

// POST /in/:endpoint_id
func (h *Handler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	endpointID := strings.TrimPrefix(r.URL.Path, "/in/")
	if endpointID == "" {
		http.Error(w, "missing endpoint id", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}

	event, err := h.store.SaveEvent(r.Context(), endpointID, r, body)
	if err != nil {
		http.Error(w, "failed to save event", http.StatusInternalServerError)
		return
	}

	targets, err := h.store.GetTargetsByEndpoint(r.Context(), endpointID)
	if err != nil {
		http.Error(w, "failed to get targets", http.StatusInternalServerError)
		return
	}

	for _, t := range targets {
		delivery, err := h.store.CreateDelivery(r.Context(), event.ID, t.ID, t.URL)
		if err != nil {
			continue
		}
		h.relay.Enqueue(delivery)
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"event_id": event.ID})
}

// GET /endpoints  |  POST /endpoints
func (h *Handler) Endpoints(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		endpoints, err := h.store.ListEndpoints(r.Context())
		if err != nil {
			http.Error(w, "failed to list endpoints", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(endpoints)

	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		ep, err := h.store.CreateEndpoint(r.Context(), body.Name)
		if err != nil {
			http.Error(w, "failed to create endpoint", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ep)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /endpoints/:id/targets
func (h *Handler) EndpointTargets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// /endpoints/:id/targets -> extract id
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[2] != "targets" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	endpointID := parts[1]

	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	target, err := h.store.CreateTarget(r.Context(), endpointID, body.URL)
	if err != nil {
		http.Error(w, "failed to create target", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}

// GET /events
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	events, err := h.store.ListEvents(r.Context())
	if err != nil {
		http.Error(w, "failed to list events", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
