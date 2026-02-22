# Sentinel

Lightweight HTTP & TLS health checker with Prometheus metrics.

I built this to solve a personal problem: I run in-cluster load balancers instead of cloud provider LBs, which means DNS records point directly at node IPs. When a node restarts and gets a new IP, things break silently. Sentinel watches my endpoints so I know when something goes unreachable and I need to update DNS.

> This is a recreation of a very old project I originally wrote in 2012. This time around, the entire thing was vibe coded.

## What it does

- Periodically hits your HTTPS endpoints and records reachability, status codes, and latency
- Checks TLS certificate validity and expiry
- Detects HTTP → HTTPS redirect presence
- Auto-discovers targets from Kubernetes Ingress resources (or uses a static config)
- Exposes everything as Prometheus metrics on `:9090/metrics`

## Quick start

### Static config

```bash
cp config.example.yaml config.yaml
# edit config.yaml with your targets
go build -o sentinel .
./sentinel
```

### Kubernetes ingress discovery

```bash
./sentinel --discover
```

In discovery mode, sentinel watches all Ingress resources across all namespaces and automatically monitors every hostname it finds. No config file needed — targets update in real time as ingresses are added, changed, or removed.

## Configuration

```yaml
interval: 5m
timeout: 10s
metrics_addr: ":9090"

targets:
  - https://app.example.com
  - https://api.example.com
  - https://cdn.example.com
```

| Field | Default | Description |
|---|---|---|
| `interval` | `5m` | Time between checks |
| `timeout` | `10s` | HTTP/TLS timeout per target |
| `metrics_addr` | `:9090` | Address for the metrics server |
| `targets` | — | List of HTTPS URLs to monitor (not needed with `--discover`) |

## Prometheus metrics

### HTTP checks

| Metric | Type | Labels | Description |
|---|---|---|---|
| `sentinel_target_up` | gauge | `host`, `url` | `1` = reachable, `0` = down |
| `sentinel_target_status_code` | gauge | `host`, `url` | Last HTTP status code |
| `sentinel_target_latency_seconds` | gauge | `host`, `url` | Check latency |
| `sentinel_checks_total` | counter | `host`, `url`, `result` | Total checks (`success`/`failure`) |

### TLS checks

| Metric | Type | Labels | Description |
|---|---|---|---|
| `sentinel_tls_redirect` | gauge | `host` | `1` if HTTP→HTTPS redirect is in place |
| `sentinel_tls_cert_valid` | gauge | `host` | `1` if certificate is valid and hostname matches |
| `sentinel_tls_cert_expiry_seconds` | gauge | `host` | Seconds until certificate expires |

### Example alerts

```yaml
# Target is down
- alert: SentinelTargetDown
  expr: sentinel_target_up == 0
  for: 10m

# Certificate expires in less than 7 days
- alert: SentinelCertExpiringSoon
  expr: sentinel_tls_cert_expiry_seconds < 604800
  for: 1h

# No HTTPS redirect
- alert: SentinelNoTLSRedirect
  expr: sentinel_tls_redirect == 0
  for: 1h
```

## Docker

```bash
docker build -t sentinel .
docker run -v ./config.yaml:/etc/sentinel/config.yaml sentinel -config /etc/sentinel/config.yaml
```

## Kubernetes

Deployment manifests are in `k8s/`. By default, sentinel runs with `--discover` and auto-discovers targets from Ingress resources.

### Ingress discovery

When running with `--discover`, sentinel uses the Kubernetes [informer](https://pkg.go.dev/k8s.io/client-go/informers) API to watch all Ingress resources across every namespace. On any ingress add, update, or delete, it rebuilds its target list by extracting every unique hostname from `spec.rules[].host` and monitoring `https://<host>`. Targets update in real time — no restarts or config reloads needed.

This means sentinel automatically picks up new services the moment an ingress is created, and stops monitoring them when the ingress is removed.

### RBAC

Sentinel needs a ServiceAccount with read-only access to Ingress resources. Here's a minimal RBAC setup:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sentinel
  namespace: apps
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: sentinel
rules:
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: sentinel
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: sentinel
subjects:
  - kind: ServiceAccount
    name: sentinel
    namespace: apps
```

A ClusterRole (not a namespaced Role) is required because sentinel watches ingresses across all namespaces. The deployment references `serviceAccountName: sentinel`.

### Deploy

```bash
kubectl apply -k k8s/
```

### CI/CD

GitHub Actions workflow builds, pushes, and deploys on every push to `main`.

**Required secrets:**

| Secret | Description |
|---|---|
| `DOCKER_REGISTRY_URL` | Container registry URL |
| `DOCKER_USERNAME` | Registry login username |
| `DOCKER_PASSWORD` | Registry login password |
| `KUBECONFIG` | Base64-encoded kubeconfig for deploy |

## License

MIT
