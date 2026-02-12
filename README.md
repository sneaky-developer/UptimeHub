# UptimeHub

A self-hosted **Kubernetes uptime monitoring system** with a public status page and admin dashboard.

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![Next.js](https://img.shields.io/badge/Next.js-15-000000?logo=next.js)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?logo=postgresql)

## Architecture

| Component | Tech | Description |
|-----------|------|-------------|
| **Master Server** | Go, Gin, GORM, PostgreSQL | Central API — stores uptime data, serves admin & status APIs |
| **Worker Agent** | Go, client-go | Runs in K8s clusters — discovers services, performs health checks |
| **Frontend** | Next.js, Tailwind CSS, React Query | Public status page + Admin dashboard |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for frontend development)

### Run with Docker Compose

```bash
# Start all services (PostgreSQL, Master, Frontend)
docker compose up --build -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

The following services will be available:
- **Frontend (Status Page)**: http://localhost:3000
- **Master API**: http://localhost:8080
- **Admin Login**: http://localhost:3000/admin/login
  - Default credentials: `admin@uptimehub.local` / `admin123`

### Local Development

```bash
# Start database only
make db-up

# Run master server
make master-run

# Run frontend dev server
make frontend-dev

# Run agent
make agent-run
```

## API Endpoints

### Public (no auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/status` | Current service statuses |
| `GET` | `/api/status/:id/history` | 90-day uptime history |
| `GET` | `/api/incidents` | Recent incidents |
| `GET` | `/api/maintenance` | Maintenance windows |

### Agent (bearer token auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/agent/register` | Register agent |
| `POST` | `/api/agent/status` | Submit check results |
| `GET` | `/api/agent/config` | Get monitoring config |
| `POST` | `/api/agent/heartbeat` | Agent heartbeat |

### Admin (JWT auth)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/admin/login` | Admin login |
| `GET/POST` | `/api/admin/services` | List/create services |
| `PUT/DELETE` | `/api/admin/services/:id` | Update/delete service |
| `GET/POST` | `/api/admin/incidents` | List/create incidents |
| `PUT` | `/api/admin/incidents/:id` | Update incident |
| `GET` | `/api/admin/agents` | List agents |

## Kubernetes Agent Setup

Label your services for auto-discovery:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-api
  labels:
    monitoring.enabled: "true"
  annotations:
    monitoring.path: "/health"
    monitoring.port: "8080"
```

## Project Structure

```
UptimeHub/
├── master/          # Go Master Server
├── agent/           # Go Worker Agent
├── frontend/        # Next.js Frontend
├── deploy/helm/     # Helm charts
├── docker-compose.yml
└── Makefile
```

## License

MIT