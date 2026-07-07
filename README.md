# UptimeHub

A self-hosted **Kubernetes uptime monitoring system** with a public status page, admin dashboard, and alerting — think of it as a status page that lives inside your clusters.

![CI](https://github.com/sneaky-developer/UptimeHub/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![Next.js](https://img.shields.io/badge/Next.js-15-000000?logo=next.js)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)
![License](https://img.shields.io/badge/License-MIT-green)

## Features

- **Public status page** — live service status, 90-day uptime history, incident timeline, maintenance windows
- **Kubernetes auto-discovery** — label a Service or Ingress and it gets monitored, no config files
- **Multi-cluster** — one master, any number of agents enrolled via group tokens
- **Auto-incidents** — services crossing their failure threshold open incidents automatically and resolve them on recovery
- **Alerting** — email, Slack, and webhook notification channels with per-channel test delivery
- **HTTP + TCP checks** — configurable interval, timeout, retries, and failure threshold per service

## Architecture

```
┌─────────────── Kubernetes cluster A ───────────────┐
│  ┌───────────┐   discovers Services/Ingresses      │
│  │   Agent   │──── labeled monitoring.enabled ──┐  │
│  │  (Go)     │                                  │  │
│  └─────┬─────┘   health-checks endpoints ◄──────┘  │
└────────┼────────────────────────────────────────────┘
         │ register / report / heartbeat (token auth)
         ▼
   ┌───────────┐      ┌────────────┐
   │  Master   │─────►│ PostgreSQL │
   │ (Go, Gin) │      └────────────┘
   └─────┬─────┘ auto-incidents ──► email / Slack / webhooks
         │ REST API
         ▼
   ┌───────────┐
   │ Frontend  │  public status page + admin dashboard
   │ (Next.js) │
   └───────────┘
```

| Component | Tech | Description |
|-----------|------|-------------|
| **Master** | Go, Gin, GORM, PostgreSQL | Central API — stores uptime data, manages incidents, sends alerts |
| **Agent** | Go, client-go | Runs in each cluster — discovers services, performs health checks |
| **Frontend** | Next.js, Tailwind CSS, React Query | Public status page + admin dashboard |

## Quick Start

### Run with Docker Compose

```bash
docker compose up --build -d
```

- **Status page**: http://localhost:3000
- **Admin dashboard**: http://localhost:3000/admin/login — `admin@uptimehub.local` / `admin123`
- **Master API**: http://localhost:8080

The bundled agent registers automatically once you create an agent group
(Admin → Agents) and set its token as `ENROLLMENT_TOKEN` in `docker-compose.yml`.

### Deploy an agent to Kubernetes

1. In the admin dashboard, create an **agent group** and copy its enrollment token.
2. Install the agent with Helm:

```bash
helm install uptimehub-agent ./deploy/helm/uptimehub-agent \
  --namespace monitoring --create-namespace \
  --set master.url=https://uptimehub.example.com \
  --set master.enrollmentToken=<token-from-step-1>
```

3. Label the services you want monitored:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-api
  labels:
    monitoring.enabled: "true"
  annotations:
    monitoring.path: "/health"   # default: /health
    monitoring.port: "8080"      # default: first service port
```

Discovered services appear under Admin → Services within a minute (private by
default — mark them public to show them on the status page). Ingresses with the
same label are monitored via their external hostname.

## Configuration

The master is configured entirely via environment variables — see
[.env.example](.env.example) for the full reference. Highlights:

| Variable | Default | Notes |
|----------|---------|-------|
| `APP_ENV` | `development` | Any other value enforces production checks |
| `JWT_SECRET` | — | **Required in production** (32+ chars, `openssl rand -hex 32`) |
| `ADMIN_EMAIL` / `ADMIN_PASSWORD` | `admin@uptimehub.local` / generated | Initial admin; production generates a random password and logs it once |
| `CHECK_RETENTION_DAYS` | `90` | Raw check results older than this are pruned daily |

Agent settings (master URL, enrollment token, check intervals, namespace scope)
are exposed as Helm values — see
[deploy/helm/uptimehub-agent/values.yaml](deploy/helm/uptimehub-agent/values.yaml).

## API Overview

### Public (no auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/status` | Current service statuses + 90-day history |
| `GET` | `/api/status/:id/history` | Per-service uptime history |
| `GET` | `/api/incidents` | Recent incidents with updates |
| `GET` | `/api/maintenance` | Upcoming/active maintenance windows |

### Agent (bearer token)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/agent/register` | Enroll with a group token, receive an agent token |
| `POST` | `/api/agent/status` | Submit batched check results |
| `POST` | `/api/agent/discovery` | Sync discovered K8s services |
| `GET` | `/api/agent/config` | Fetch assigned services and check settings |
| `POST` | `/api/agent/heartbeat` | Liveness ping |

### Admin (JWT)
CRUD for services, incidents, maintenance windows, agent groups, and alert
channels under `/api/admin/*` — see [master/cmd/server/main.go](master/cmd/server/main.go)
for the full route list.

## Local Development

```bash
make db-up          # PostgreSQL only
make master-run     # API on :8080
make frontend-dev   # UI on :3000
make agent-run      # needs ENROLLMENT_TOKEN + MASTER_URL

make master-test agent-test   # run test suites
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Project Structure

```
UptimeHub/
├── master/                      # Go master server (API, alerting, aggregation)
├── agent/                       # Go worker agent (discovery, health checks)
├── frontend/                    # Next.js status page + admin dashboard
├── deploy/helm/uptimehub-agent/ # Helm chart for the in-cluster agent
├── .github/workflows/           # CI: build, test, lint, Docker images
└── docker-compose.yml           # Full local stack
```

## Roadmap

- [ ] Prometheus `/metrics` endpoint on master and agent
- [ ] OpenAPI specification
- [ ] Helm chart for the master + frontend
- [ ] Maintenance window management UI
- [ ] Multi-user admin with roles

## License

[MIT](LICENSE)
