package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ahmethamdiozen/webhook-relay/internal/handler"
	"github.com/ahmethamdiozen/webhook-relay/internal/relay"
	"github.com/ahmethamdiozen/webhook-relay/internal/store"
)

func main() {
	ctx := context.Background()

	db, err := store.Connect(ctx)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("connected to database")

	s := store.New(db)
	r := relay.New(s, 5) // 5 worker goroutine
	h := handler.New(s, r)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	h.RegisterRoutes(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
