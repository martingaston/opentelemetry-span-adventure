package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Package-level tracer.
// This should be configured in your code setup instead of here.
var tracer = otel.Tracer("github.com/full/path/to/mypkg")

// getUser emulates getting a user from the db
func getUser(ctx context.Context) {
	_, span := tracer.Start(ctx, "getUser")
	defer span.End()

	id := rand.Intn(10) + 1

	span.SetAttributes(attribute.Int("user", id))
}

// getOrder emulates getting an order from the db
func getOrder(ctx context.Context) {
	_, span := tracer.Start(ctx, "getOrder")
	defer span.End()

	id := uuid.New().String()

	span.SetAttributes(attribute.String("order", id))
}

// goBoom emulates a big nasty error
func goBoom() bool {
	unluckyNumber := rand.Intn(5)

	return unluckyNumber == 0
}

// sleepy mocks work that your application does.
func sleepy(ctx context.Context) {
	_, span := tracer.Start(ctx, "sleep")
	defer span.End()

	sleepTime := 1 * time.Second
	time.Sleep(sleepTime)
	span.SetAttributes(attribute.Int("sleep.duration", int(sleepTime)))
}

// httpHandler is an HTTP handler function that is going to be instrumented.
func httpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	getUser(ctx)
	getOrder(ctx)
	sleepy(ctx)
	if goBoom() {
		w.WriteHeader(http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "Hello, World! I am instrumented automatically!")
}

func main() {
	ctx := context.Background()
	tp := traceWithHoneycomb(ctx)
	defer func() { _ = tp.Shutdown(ctx) }()

	// Wrap your httpHandler function.
	handler := http.HandlerFunc(httpHandler)
	wrappedHandler := otelhttp.NewHandler(handler, "hello-instrumented")
	http.Handle("/hello-instrumented", wrappedHandler)

	// And start the HTTP serve.
	log.Fatal(http.ListenAndServe(":3030", nil))
}
