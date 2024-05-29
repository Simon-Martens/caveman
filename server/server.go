package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Simon-Martens/caveman/app"
	"github.com/labstack/echo/v4"
)

// INFO:Server structure inspired by
// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/
func Serve(ctx context.Context, app *app.App) error {
	mux := NewMux(app)
	AddRoutes(mux, app)
	httpServer := NewServer(mux)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

func NewServer(mux *echo.Echo) *http.Server {
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return httpServer
}
