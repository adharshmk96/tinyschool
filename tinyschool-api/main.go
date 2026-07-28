package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tinyschool-api/internal/httpapi"
)

func main() {
	defaultAddress := os.Getenv("TINYSCHOOL_API_ADDRESS")
	if defaultAddress == "" {
		defaultAddress = ":8080"
	}
	address := flag.String("address", defaultAddress, "HTTP listen address")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := httpapi.NewServer(*address, httpapi.NewHandler(logger))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		logger.Info("shutting down API")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("Tiny School API listening", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("API stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
