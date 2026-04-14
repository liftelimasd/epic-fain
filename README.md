# Epic Fain

A robust IoT platform for real-time telemetry collection and device management via CAN bus communication. Built with Go and designed for scalability, reliability, and compliance monitoring of industrial equipment.

## Overview

Epic Fain is a microservices-ready backend application that collects, processes, and manages telemetry data from Epic-series industrial devices. It handles CAN bus frame parsing, real-time monitoring, audit logging, and alert generation with a clean hexagonal architecture.

### Key Features

- **CAN Bus Integration**: Real-time reception and parsing of CAN frames from remote devices
- **Telemetry Management**: Store and retrieve device measurements including voltage, current, power, and lithium battery data
- **Audit Logging**: Complete audit trail of all system operations for compliance
- **Alert System**: Automatic alert generation for anomalies and threshold violations
- **Device Control**: Command devices through the platform (Hito 3)
- **REST API**: Comprehensive HTTP API for system interaction
- **Multi-Protocol**: Dual communication via HTTP and TCP
- **Database Persistence**: PostgreSQL for reliable data storage and retention policies

## Technology Stack

- **Language**: Go 1.22
- **Database**: PostgreSQL 16
- **Containerization**: Docker & Docker Compose
- **Architecture**: Hexagonal (Ports & Adapters)
- **Key Dependencies**:
  - `github.com/lib/pq` - PostgreSQL driver
  - `github.com/google/uuid` - UUID generation

## Architecture

The project follows a **Hexagonal Architecture** (Ports & Adapters) pattern:

```
epic-fain/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── domain/          # Core business logic (independent of infrastructure)
│   │   ├── model/       # Domain entities (CANFrame, Telemetry, Alert, Audit, etc.)
│   │   ├── port/        # Interface definitions (Repository, Event, etc.)
│   │   └── service/     # Domain services (CAN decoder, business rules)
│   ├── application/     # Application services orchestrating domain logic
│   │   ├── device_control_service.go
│   │   └── telemetry_service.go
│   └── infrastructure/  # External adapters
│       ├── adapter/
│       │   ├── inbound/     # HTTP & TCP servers
│       │   └── outbound/    # Database repositories, external services
│       ├── config/          # Configuration management
│       └── migration/       # Database schemas
```

### Domain Model

Key entities managed by the platform:

- **CANFrame**: Raw CAN bus messages with MessageID, DLC, and data
- **Telemetry**: Device measurements (status, voltage, current, power, lithium data)
- **Command**: Control directives sent to devices (VVVF control, measurement config)
- **Alert**: Generated alerts for anomalies and threshold breaches
- **AuditLog**: Complete record of platform operations
- **Installation**: Device/installation configuration and metadata

### CAN Message Protocol

#### Inbound (Devices → Platform)
- `0xFF00` - Status (7 bytes)
- `0xFF01` - Measurements (8 bytes)
- `0xFF02` - Currents (4 bytes)
- `0xFF03` - Lithium data (8 bytes, lithium-enabled devices only)
- `0xFF0E` - Device info (8 bytes)

#### Outbound (Platform → Devices)
- `0xEF00` - VVVF/Reset control (1 byte)
- `0xEF01` - Measurement configuration (3 bytes)
- `0xEF0E` - Device info request (1 byte)

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16+ (or use Docker)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/liftelimasd/epic-fain.git
   cd epic-fain
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start with Docker Compose** (recommended)
   ```bash
   docker-compose up
   ```
   This starts both PostgreSQL and the Epic Fain application.

4. **Or run locally**
   ```bash
   # Ensure PostgreSQL is running
   # Set environment variables
   export HTTP_ADDR=":8080"
   export TCP_ADDR=":9090"
   export DB_HOST="localhost"
   export DB_PORT="5432"
   export DB_USER="epicfain"
   export DB_PASSWORD="epicfain"
   export DB_NAME="epicfain"
   export DB_SSLMODE="disable"
   
   go run ./cmd/server
   ```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_ADDR` | `:8080` | HTTP server listening address |
| `TCP_ADDR` | `:9090` | TCP server listening address (CAN data receiver) |
| `DB_HOST` | `localhost` | PostgreSQL hostname |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `epicfain` | PostgreSQL username |
| `DB_PASSWORD` | `epicfain` | PostgreSQL password |
| `DB_NAME` | `epicfain` | PostgreSQL database name |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `API_KEYS` | (optional) | Comma-separated API keys for authentication |

## API Endpoints

All HTTP endpoints require API key authentication (header: `X-API-Key`).

### Telemetry
- `GET /api/v1/telemetry` - Retrieve telemetry data
- `POST /api/v1/telemetry` - Create new telemetry record

### Audit
- `GET /api/v1/audit` - Retrieve audit logs

### Devices
- `GET /api/v1/devices` - List devices (Hito 2+)
- `POST /api/v1/devices/{id}/command` - Send command to device (Hito 3)

### Health & Info
- `GET /health` - Health check endpoint
- `GET /api/v1/info` - Platform information

## TCP Protocol

The platform listens on the TCP port for CAN frame data. Devices should send raw CAN frames in the following binary format:

```
[MessageID: 2 bytes][DLC: 1 byte][Data: up to 8 bytes]
```

## Database Schema

Automatic migrations run on startup:

1. **001_initial_schema.sql** - Tables for telemetry, audit, alerts, installations, commands
2. **002_retention_policy.sql** - Data retention rules and policies

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

Key test files:
- `internal/domain/service/decoder_test.go` - CAN decoder tests

## Development

### Code Style

Follow Go conventions and idioms:
- Use table-driven tests
- Implement proper error handling
- Keep packages focused and modular
- Respect the architecture boundaries

### Recommended Skills

The project includes automated Go development skills:
- **Go Development Patterns** - Idiomatic patterns and best practices
- **Go Testing Patterns** - TDD methodology with table-driven tests

Access them in `.claude/skills/` or `.agents/skills/`.

### Project Phases

- **Hito 1**: Core telemetry collection and API ✅
- **Hito 2**: Device management, installation tracking
- **Hito 3**: Alert system, device control, MQTT integration

## Deployment

### Docker

1. **Build the image**
   ```bash
   docker build -t liftel/epic-fain:latest .
   ```

2. **Run with compose**
   ```bash
   docker-compose up -d
   ```

3. **View logs**
   ```bash
   docker-compose logs -f epic-fain
   ```

4. **Stop services**
   ```bash
   docker-compose down
   ```

### Health Check

The application exposes a health endpoint:
```bash
curl http://localhost:8080/health
```

## Database Migrations

Migrations are applied automatically on startup from SQL files in `internal/infrastructure/migration/`.

To manually run migrations:
```bash
# Ensure PostgreSQL CLI tools are installed
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < internal/infrastructure/migration/001_initial_schema.sql
psql -h $DB_HOST -U $DB_USER -d $DB_NAME < internal/infrastructure/migration/002_retention_policy.sql
```

## Troubleshooting

### Database connection errors
- Verify PostgreSQL is running: `docker-compose ps`
- Check database credentials match environment variables
- Ensure database is healthy: `docker-compose logs db`

### TCP server not receiving data
- Verify TCP_ADDR is correctly configured
- Check firewall rules allow connections to the TCP port
- Ensure remote devices are sending to the correct host/port

### API authentication failures
- Verify `API_KEYS` environment variable is set
- Ensure `X-API-Key` header is included in requests
- Check API key format (comma-separated if multiple)

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Commit your changes with clear messages
4. Push to your fork
5. Submit a pull request

## License

[Specify your license here - e.g., MIT, Apache 2.0, etc.]

## Support

For issues, questions, or feedback:
- Open an issue on GitHub: https://github.com/liftelimasd/epic-fain/issues
- Contact the Liftel team

## Project Information

- **Repository**: https://github.com/liftelimasd/epic-fain
- **Organization**: [Liftel](https://github.com/liftelimasd)
- **Language**: Go 1.22
- **Last Updated**: 2026
