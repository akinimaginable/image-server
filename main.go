package main

import (
	"context"
	"errors"
	"image-server/internal/images"
	"image-server/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	imageDir := env("IMAGE_DIR", "./images")
	port := env("PORT", "8080")
	rescanInterval := 12 * time.Hour

	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", imageDir)
	}

	repo := images.NewFSRepository(imageDir, rescanInterval)
	imageServer := server.New(repo, imageDir)

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      imageServer.Handler(),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	go func() {
		log.Printf("Serving %s on http://localhost%s, ", imageDir, httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http httpServer: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Server stopped gracefully")
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
