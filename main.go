package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	listenAddr      = "127.0.0.1:4545"
	shutDownTimeout = 3 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := runServer(ctx); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx context.Context) error {
	var (
		mux = echo.New()
		srv = http.Server{
			Addr:    listenAddr,
			Handler: mux,
		}
	)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	log.Printf("listening on %s", listenAddr)
	<-ctx.Done()

	log.Println("shutdown gracefully")

	shutDownWithCtx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutDownWithCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	longShutdown := make(chan struct{}, 1)
	go func() {
		time.Sleep(3 * time.Second)
		longShutdown <- struct{}{}
	}()

	select {
	case <-shutDownWithCtx.Done():
		return fmt.Errorf("shutdown :%w", ctx.Err())
	case <-longShutdown:
		log.Println("finished")
	}
	return nil
}
