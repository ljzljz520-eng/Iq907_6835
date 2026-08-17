package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"volunteertraining/internal/app"
	"volunteertraining/internal/httpapi"
)

func main() {
	databasePath := flag.String("db", "volunteer-training.db", "BoltDB path")
	address := flag.String("addr", ":8080", "HTTP listen address")
	seed := flag.Bool("seed", false, "seed deterministic fixtures")
	flag.Parse()
	application, err := app.Open(*databasePath, time.Now)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer application.Close()
	if *seed {
		if err := application.SeedFixtures(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	server := &http.Server{Addr: *address, Handler: httpapi.New(application).Handler(), ReadHeaderTimeout: 5 * time.Second}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
