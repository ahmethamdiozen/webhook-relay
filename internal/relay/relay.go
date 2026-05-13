package relay

import (
	"context"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/ahmethamdiozen/webhook-relay/internal/model"
	"github.com/ahmethamdiozen/webhook-relay/internal/store"
)

type Job struct {
	Delivery *model.Delivery
}

type Relay struct {
	store  *store.Store
	queue  chan Job
	client *http.Client
}

func New(s *store.Store, workers int) *Relay {
	r := &Relay{
		store:  s,
		queue:  make(chan Job, 512),
		client: &http.Client{Timeout: 10 * time.Second},
	}
	for i := 0; i < workers; i++ {
		go r.worker()
	}
	return r
}

func (r *Relay) Enqueue(d *model.Delivery) {
	r.queue <- Job{Delivery: d}
}

func (r *Relay) worker() {
	for job := range r.queue {
		r.deliver(job.Delivery)
	}
}

func (r *Relay) deliver(d *model.Delivery) {
	const maxAttempts = 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := r.send(d.TargetURL, d.EventID)
		d.Attempts = attempt

		if err == nil {
			_ = r.store.UpdateDelivery(context.Background(), d.ID, "success", "", attempt)
			log.Printf("delivery %s succeeded on attempt %d", d.ID, attempt)
			return
		}

		log.Printf("delivery %s attempt %d failed: %v", d.ID, attempt, err)

		if attempt < maxAttempts {
			// exponential backoff: 2s, 4s, 8s, 16s
			wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(wait)
		}
	}

	_ = r.store.UpdateDelivery(context.Background(), d.ID, "failed", "max attempts reached", maxAttempts)
}

func (r *Relay) send(targetURL, eventID string) error {
	resp, err := r.client.Post(targetURL, "application/json", strings.NewReader(`{"event_id":"`+eventID+`"}`))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
