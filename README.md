# webhook-relay

A lightweight webhook relay service written in Go. Receives incoming webhooks, queues them, and forwards to multiple destinations with retry logic.

## What it does

- Receives webhooks at `/in/:endpoint_id`
- Stores events in PostgreSQL
- Forwards to configured destinations via a worker pool (goroutines)
- Retries failed deliveries with exponential backoff
- REST API for managing endpoints and viewing event history

## Stack

- Go (standard library `net/http`)
- PostgreSQL
- Docker

## Getting Started

```bash
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

## API

```
POST   /in/:endpoint_id          # receive a webhook
POST   /endpoints                # create endpoint
GET    /endpoints                # list endpoints
POST   /endpoints/:id/targets    # add delivery target
GET    /events                   # list events
GET    /events/:id               # event detail with delivery status
```
