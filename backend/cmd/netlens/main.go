// Command netlens is the NetLens backend: a single binary serving the REST
// API, WebSocket events, and (once embedded) the SvelteKit frontend.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netlens/backend/internal/api"
	"netlens/backend/internal/config"
)

// Set at build time via -ldflags "-X main.version=... -X main.revision=...".
var (
	version  = "dev"
	revision = "unknown"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatalf("netlens: %v", err)
	}
}

func run(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	handler := api.NewHandler(api.Status{App: "netlens", Version: version, Revision: revision})
	srv := &http.Server{Handler: handler}

	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("netlens %s (%s) listening on http://%s", version, revision, ln.Addr())
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Println("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	}
}
