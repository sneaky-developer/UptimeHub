# Contributing to UptimeHub

Thanks for your interest in contributing! This document covers the basics.

## Development setup

Prerequisites: Docker, Go 1.22+, Node.js 20+.

```bash
# Start everything
make dev

# Or run pieces individually against a local database
make db-up
make master-run       # in one terminal
make frontend-dev     # in another
make agent-run        # optional; needs ENROLLMENT_TOKEN (create a group in the admin UI)
```

The admin UI is at http://localhost:3000/admin/login (dev credentials: `admin@uptimehub.local` / `admin123`).

## Before opening a PR

```bash
# Go components
make master-test
make agent-test
cd master && go vet ./...
cd agent && go vet ./...

# Frontend
cd frontend && npm run build

# Helm chart (if you touched it)
helm lint deploy/helm/uptimehub-agent \
  --set master.url=https://example.com --set master.enrollmentToken=x
```

CI runs the same checks on every pull request.

## Guidelines

- Keep PRs focused — one feature or fix per PR.
- Add or update tests for behavior changes.
- New master configuration must come from environment variables with sane
  defaults (see `master/internal/config/config.go`) and be documented in
  `.env.example`.
- Database schema changes go through GORM models; migrations run
  automatically via AutoMigrate.

## Reporting bugs

Open an issue with steps to reproduce, expected vs. actual behavior, and
relevant logs (`make dev-logs`).
