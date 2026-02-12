# Kubernetes Uptime Monitoring System

## 1. Purpose (In Simple Terms)

We are building a self-hosted system to:

- Monitor uptime of microservices and APIs running on Kubernetes
- Monitor any custom HTTP/HTTPS endpoint
- Show service status on a public status page
- Allow admins to manage monitoring from one place

Main focus: **Is the service up or down?**

---

## 2. What We Are Building

The system has two main parts:

### 2.1 Master Server

- Central server
- Receives health data from clusters
- Stores uptime history
- Shows public status page
- Provides admin panel

### 2.2 Worker Agent (Runs in Kubernetes)

- Runs inside each Kubernetes cluster
- Checks service endpoints
- Sends status to Master
- Follows configuration from Master

---

## 3. High-Level Flow

1. Worker runs in cluster
2. Discovers services/endpoints
3. Performs health checks
4. Sends results to Master
5. Master updates status page
6. Admin manages config from dashboard

---

## 4. Scope (What We Will Monitor)

### 4.1 Primary (Main Goal)

- HTTP/HTTPS endpoint availability
- Response status (200 / non-200)
- Response time
- Timeout / connection failures

### 4.2 Secondary (For Debugging Only)

- Pod restarts
- CrashLoopBackOff
- Basic CPU/Memory (optional)

These are only used to help understand failures.

---

## 5. Master Server Requirements

### 5.1 Core Features

- Receive uptime data from agents
- Store check results
- Show service status
- Incident history
- Maintenance notices

### 5.2 Admin Panel

- Login: /api/admin/login
- Manage agents
- Add/edit monitored endpoints
- Configure intervals and retries
- Create incidents manually

### 5.3 Public Status Page

- URL: [https://status.domain.com](https://status.domain.com)
- Read-only
- Shows:
  - Current status
  - Past incidents
  - Uptime history

---

## 6. Worker Agent Requirements

### 6.1 Deployment

- Runs as Kubernetes Deployment
- Installed via Helm
- Uses ConfigMaps + Secrets
- One agent per cluster (default)

### 6.2 Service Discovery

- Uses Kubernetes API
- Finds Services/Ingresses with labels

Example:
monitoring.enabled: "true"
monitoring.path: "/health"

---

### 6.3 Health Checks

Supports:

- HTTP/HTTPS checks
- TCP checks (optional)
- Custom URLs

Configurable:

- Interval
- Timeout
- Retries
- Failure threshold

---

### 6.4 Agent Behavior

- Auto-registers with Master
- Fetches config periodically
- Reloads config without restart
- Sends heartbeat
- Reconnects automatically

---

## 7. Security

- All traffic over TLS
- Agent authentication via tokens
- Credentials stored in Kubernetes Secrets
- RBAC for admin users

---

## 8. Performance Targets

- Must handle 100+ services per cluster
- Detect failure within 30 seconds
- Minimal resource usage

Target:

- CPU < 200m
- Memory < 150MB

---

## 9. APIs (Minimum Set)

### Agent APIs

- POST /api/agent/register
- POST /api/agent/status
- GET /api/agent/config
- POST /api/agent/heartbeat

### Admin APIs

- POST /api/admin/login
- GET /api/admin/services
- PUT /api/admin/config
- POST /api/admin/incidents

---

## 10. Deployment Model

- Containerized (Docker)
- Kubernetes (GKE)
- Ingress + TLS
- Helm charts
- cert-manager for certificates

---

## 11. Finalized Technology Stack (Approved)

### Backend (Master)

- Language: Go
- Framework: Gin
- API: REST + OpenAPI
- Auth: JWT

### Worker Agent

- Language: Go
- Kubernetes Client: client-go
- Container: Distroless

### Database

- PostgreSQL (main DB)
- Redis (cache, optional)

### Frontend

- Next.js (React)
- Tailwind CSS
- React Query

### Observability

- Prometheus (metrics)
- Grafana (dashboards)
- Loki (logs)

### CI/CD & Infra

- GitHub Actions
- ArgoCD
- Helm
- Terraform
- GCP Artifact Registry

---

## 12. Phases

### Phase 1 (MVP)

- Agent + Master
- HTTP monitoring
- Status page
- Basic admin panel

### Phase 2

- Alerting (Slack/Email)
- HA setup
- Advanced RBAC

##
