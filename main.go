package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arijitdasgupta/sentinel/internal/checker"
	"github.com/arijitdasgupta/sentinel/internal/config"
	"github.com/arijitdasgupta/sentinel/internal/discovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	discover := flag.Bool("discover", false, "discover targets from kubernetes ingresses")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	cfg, err := config.Load(*configPath)
	if err != nil && !*discover {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	if cfg == nil {
		cfg = config.Default()
	}

	slog.Info("sentinel starting",
		"discover", *discover,
		"targets", len(cfg.Targets),
		"interval", cfg.Interval,
		"metrics_addr", cfg.MetricsAddr,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	metricsServer := &http.Server{
		Addr:    cfg.MetricsAddr,
		Handler: mux,
	}

	go func() {
		slog.Info("metrics server listening", "addr", cfg.MetricsAddr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", "error", err)
			os.Exit(1)
		}
	}()

	c := checker.New(cfg)

	if *discover {
		_, err := discovery.New(func(targets []config.Target) {
			c.UpdateTargets(targets)
		})
		if err != nil {
			slog.Error("failed to start ingress watcher", "error", err)
			os.Exit(1)
		}
	}

	go c.Run(ctx)

	<-ctx.Done()
	slog.Info("shutting down")

	metricsServer.Close()
}
