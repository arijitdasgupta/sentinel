package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/arijitdasgupta/sentinel/internal/config"
	"github.com/arijitdasgupta/sentinel/internal/metrics"
)

type Checker struct {
	mu               sync.RWMutex
	targets          []config.Target
	client           *http.Client
	noRedirectClient *http.Client
	interval         time.Duration
	timeout          time.Duration
}

func New(cfg *config.Config) *Checker {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:  cfg.Timeout,
		ResponseHeaderTimeout: cfg.Timeout,
		DisableKeepAlives:    true,
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	noRedirectClient := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Checker{
		targets:          cfg.Targets,
		client:           client,
		noRedirectClient: noRedirectClient,
		interval:         cfg.Interval,
		timeout:          cfg.Timeout,
	}
}

func (c *Checker) UpdateTargets(targets []config.Target) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.targets = targets
	slog.Info("checker targets updated", "count", len(targets))
}

func (c *Checker) Run(ctx context.Context) {
	c.mu.RLock()
	count := len(c.targets)
	c.mu.RUnlock()
	slog.Info("starting checker", "targets", count, "interval", c.interval)

	c.checkAll(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("checker stopped")
			return
		case <-ticker.C:
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	c.mu.RLock()
	targets := make([]config.Target, len(c.targets))
	copy(targets, c.targets)
	c.mu.RUnlock()

	slog.Info("running checks", "targets", len(targets))

	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(t config.Target) {
			defer wg.Done()
			c.check(ctx, t)
			c.checkTLS(t)
		}(t)
	}
	wg.Wait()
}

func (c *Checker) check(ctx context.Context, t config.Target) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		slog.Error("creating request", "host", t.Host, "error", err)
		c.recordDown(t, 0, time.Since(start))
		return
	}

	req.Header.Set("User-Agent", "sentinel-health-checker/1.0")

	resp, err := c.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		slog.Warn("target unreachable", "host", t.Host, "url", t.URL, "error", err)
		c.recordDown(t, 0, latency)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		slog.Info("target up", "host", t.Host, "status", resp.StatusCode, "latency", latency)
		c.recordUp(t, resp.StatusCode, latency)
	} else {
		slog.Warn("target returned error", "host", t.Host, "status", resp.StatusCode, "latency", latency)
		c.recordDown(t, resp.StatusCode, latency)
	}
}

func (c *Checker) checkTLS(t config.Target) {
	c.checkTLSRedirect(t)
	c.checkCertificate(t)
}

func (c *Checker) checkTLSRedirect(t config.Target) {
	httpURL := fmt.Sprintf("http://%s/", t.Host)
	resp, err := c.noRedirectClient.Get(httpURL)
	if err != nil {
		slog.Debug("tls redirect check failed", "host", t.Host, "error", err)
		metrics.TLSRedirect.WithLabelValues(t.Host).Set(0)
		return
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	if resp.StatusCode >= 300 && resp.StatusCode < 400 && strings.HasPrefix(location, "https://") {
		slog.Info("tls redirect active", "host", t.Host)
		metrics.TLSRedirect.WithLabelValues(t.Host).Set(1)
	} else {
		slog.Warn("no tls redirect", "host", t.Host, "status", resp.StatusCode)
		metrics.TLSRedirect.WithLabelValues(t.Host).Set(0)
	}
}

func (c *Checker) checkCertificate(t config.Target) {
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: c.timeout},
		"tcp",
		t.Host+":443",
		&tls.Config{ServerName: t.Host},
	)
	if err != nil {
		slog.Warn("tls handshake failed", "host", t.Host, "error", err)
		metrics.TLSCertValid.WithLabelValues(t.Host).Set(0)
		metrics.TLSCertExpirySeconds.WithLabelValues(t.Host).Set(0)
		return
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		slog.Warn("no peer certificates", "host", t.Host)
		metrics.TLSCertValid.WithLabelValues(t.Host).Set(0)
		metrics.TLSCertExpirySeconds.WithLabelValues(t.Host).Set(0)
		return
	}

	leaf := certs[0]
	now := time.Now()
	expiresIn := leaf.NotAfter.Sub(now).Seconds()

	metrics.TLSCertExpirySeconds.WithLabelValues(t.Host).Set(expiresIn)

	if now.Before(leaf.NotBefore) || now.After(leaf.NotAfter) {
		slog.Warn("certificate not valid", "host", t.Host, "notBefore", leaf.NotBefore, "notAfter", leaf.NotAfter)
		metrics.TLSCertValid.WithLabelValues(t.Host).Set(0)
	} else {
		err := leaf.VerifyHostname(t.Host)
		if err != nil {
			slog.Warn("certificate hostname mismatch", "host", t.Host, "error", err)
			metrics.TLSCertValid.WithLabelValues(t.Host).Set(0)
		} else {
			slog.Info("certificate valid", "host", t.Host, "expires_in_days", int(expiresIn/86400))
			metrics.TLSCertValid.WithLabelValues(t.Host).Set(1)
		}
	}
}

func (c *Checker) recordUp(t config.Target, status int, latency time.Duration) {
	labels := []string{t.Host, t.URL}
	metrics.TargetUp.WithLabelValues(labels...).Set(1)
	metrics.TargetStatusCode.WithLabelValues(labels...).Set(float64(status))
	metrics.TargetLatencySeconds.WithLabelValues(labels...).Set(latency.Seconds())
	metrics.CheckTotal.WithLabelValues(append(labels, "success")...).Inc()
}

func (c *Checker) recordDown(t config.Target, status int, latency time.Duration) {
	labels := []string{t.Host, t.URL}
	metrics.TargetUp.WithLabelValues(labels...).Set(0)
	metrics.TargetStatusCode.WithLabelValues(labels...).Set(float64(status))
	metrics.TargetLatencySeconds.WithLabelValues(labels...).Set(latency.Seconds())
	metrics.CheckTotal.WithLabelValues(append(labels, "failure")...).Inc()
}
