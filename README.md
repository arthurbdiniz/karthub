# KartHub

Self-hosted web application for amateur karting groups. Organize races, manage championships, book seats, track statistics.

## Quick Start

```bash
# Development
make dev

# Docker
docker compose up -d
```

The app runs at http://localhost:8080.

## Features

- Event scheduling and booking (with waitlist)
- Championship management with configurable points systems
- Driver profiles and statistics
- Race result tracking with automatic points calculation
- Leaderboards and rankings
- Admin interface for CRUD operations
- Mobile-friendly responsive UI

## Tech Stack

- Go 1.24+ with Chi router
- SQLite (WAL mode) with Goose migrations
- HTMX + Alpine.js for interactivity
- Tailwind CSS for styling
- Cookie-based sessions

## Configuration

Copy `config.yaml` and edit, or use environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `KARTHUB_PORT` | Server port | 8080 |
| `KARTHUB_DB_PATH` | SQLite database path | data/karthub.db |
| `KARTHUB_SESSION_SECRET` | Session cookie secret | change-me |
| `KARTHUB_LOG_LEVEL` | Log level (debug/info/warn/error) | info |

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make fmt      # Format code
```

## Deployment

```bash
docker compose up -d
```

Data persists in the `karthub_data` Docker volume.

## License

MIT
