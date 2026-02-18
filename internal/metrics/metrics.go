package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TargetUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_target_up",
		Help: "Whether the target is reachable (1 = up, 0 = down).",
	}, []string{"host", "url"})

	TargetStatusCode = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_target_status_code",
		Help: "HTTP status code returned by the target.",
	}, []string{"host", "url"})

	TargetLatencySeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_target_latency_seconds",
		Help: "Latency of the HTTP check in seconds.",
	}, []string{"host", "url"})

	CheckTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_checks_total",
		Help: "Total number of checks performed.",
	}, []string{"host", "url", "result"})

	TLSRedirect = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_tls_redirect",
		Help: "Whether HTTP redirects to HTTPS (1 = yes, 0 = no).",
	}, []string{"host"})

	TLSCertExpirySeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_tls_cert_expiry_seconds",
		Help: "Seconds until the TLS certificate expires.",
	}, []string{"host"})

	TLSCertValid = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_tls_cert_valid",
		Help: "Whether the TLS certificate is valid (1 = valid, 0 = invalid).",
	}, []string{"host"})
)
