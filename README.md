# Ticket Service

A comprehensive ticketing/helpdesk microservice for the Minisource platform.

## Features

### Core Features
- **Multi-tenant Support**: Full tenant isolation for all data
- **Ticket Management**: Create, update, assign, transfer, and close tickets
- **Department & Category Organization**: Hierarchical departments and categories
- **SLA Management**: Configurable SLA policies with priority-based response/resolution times
- **Agent Management**: Agent roles, skills, availability, and workload management
- **Canned Responses**: Pre-defined responses for common queries

### Ticket Features
- Multiple ticket sources (web, email, phone, chat, API, social, widget)
- Multiple ticket types (question, incident, problem, feature request, task)
- Priority levels (low, medium, high, urgent, critical)
- Status workflow (open → in progress → pending → resolved → closed)
- Auto-assignment (round-robin, load-balanced, skill-based)
- Ticket transfer between departments
- Customer satisfaction rating
- File attachments
- Custom fields
- Tags and labels
- Related tickets
- Watchers

### Communication
- Customer replies
- Agent replies
- Internal notes (private)
- System messages
- Auto-replies

### Admin Features
- Dashboard with statistics
- Bulk operations (assign, status change, priority change, transfer, delete)
- SLA breach monitoring
- Due soon tickets
- Unassigned tickets queue
- Agent performance metrics

## API Endpoints

### Health
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /live` - Liveness check

### Tickets (Customer/User)
- `POST /api/v1/tickets` - Create ticket
- `GET /api/v1/tickets` - List tickets
- `GET /api/v1/tickets/:id` - Get ticket
- `GET /api/v1/tickets/number/:number` - Get ticket by number
- `PATCH /api/v1/tickets/:id` - Update ticket
- `DELETE /api/v1/tickets/:id` - Delete ticket
- `PATCH /api/v1/tickets/:id/status` - Change status
- `POST /api/v1/tickets/:id/rate` - Rate ticket
- `GET /api/v1/tickets/:id/messages` - Get messages
- `POST /api/v1/tickets/:id/messages` - Add message
- `GET /api/v1/tickets/:id/history` - Get history
- `GET /api/v1/tickets/stats` - Get statistics
- `GET /api/v1/customers/:customer_id/tickets` - Get customer tickets

### Tickets (Agent)
- `POST /api/v1/tickets/:id/assign` - Assign ticket
- `POST /api/v1/tickets/:id/transfer` - Transfer ticket
- `GET /api/v1/agents/:agent_id/tickets` - Get agent tickets

### Admin - Agents
- `POST /api/v1/admin/agents` - Create agent
- `GET /api/v1/admin/agents` - List agents
- `GET /api/v1/admin/agents/:id` - Get agent
- `PATCH /api/v1/admin/agents/:id` - Update agent
- `DELETE /api/v1/admin/agents/:id` - Delete agent
- `PATCH /api/v1/admin/agents/:id/status` - Update agent status

### Admin - Departments
- `POST /api/v1/admin/departments` - Create department
- `GET /api/v1/admin/departments` - List departments
- `GET /api/v1/admin/departments/:id` - Get department
- `PATCH /api/v1/admin/departments/:id` - Update department
- `DELETE /api/v1/admin/departments/:id` - Delete department
- `GET /api/v1/admin/departments/:id/agents` - Get department agents
- `POST /api/v1/admin/departments/:id/agents` - Add agent to department
- `DELETE /api/v1/admin/departments/:id/agents/:agent_id` - Remove agent from department

### Admin - Categories
- `POST /api/v1/admin/categories` - Create category
- `GET /api/v1/admin/categories` - List categories
- `GET /api/v1/admin/categories/:id` - Get category
- `PATCH /api/v1/admin/categories/:id` - Update category
- `DELETE /api/v1/admin/categories/:id` - Delete category

### Admin - SLA Policies
- `POST /api/v1/admin/sla-policies` - Create SLA policy
- `GET /api/v1/admin/sla-policies` - List SLA policies
- `GET /api/v1/admin/sla-policies/:id` - Get SLA policy
- `PATCH /api/v1/admin/sla-policies/:id` - Update SLA policy
- `DELETE /api/v1/admin/sla-policies/:id` - Delete SLA policy

### Admin - Canned Responses
- `POST /api/v1/admin/canned-responses` - Create canned response
- `GET /api/v1/admin/canned-responses` - List canned responses
- `GET /api/v1/admin/canned-responses/search` - Search canned responses
- `GET /api/v1/admin/canned-responses/:id` - Get canned response
- `PATCH /api/v1/admin/canned-responses/:id` - Update canned response
- `DELETE /api/v1/admin/canned-responses/:id` - Delete canned response

### Admin - Bulk Operations
- `POST /api/v1/admin/tickets/bulk-assign` - Bulk assign tickets
- `POST /api/v1/admin/tickets/bulk-status` - Bulk change status
- `POST /api/v1/admin/tickets/bulk-priority` - Bulk change priority
- `POST /api/v1/admin/tickets/bulk-transfer` - Bulk transfer department
- `POST /api/v1/admin/tickets/bulk-delete` - Bulk delete tickets

### Admin - Dashboard
- `GET /api/v1/admin/dashboard/stats` - Get dashboard statistics
- `GET /api/v1/admin/dashboard/sla-breached` - Get SLA breached tickets
- `GET /api/v1/admin/dashboard/due-soon` - Get tickets due soon
- `GET /api/v1/admin/dashboard/unassigned` - Get unassigned tickets

## Configuration

Environment variables:

```env
# Server
SERVER_PORT=5011

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=ticket_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Auth Service
AUTH_URL=http://localhost:5001
AUTH_CLIENT_ID=ticket-service
AUTH_CLIENT_SECRET=your-secret

# Notifier Service
NOTIFIER_URL=http://localhost:5003
NOTIFIER_ENABLED=true

# SLA Defaults (in hours)
SLA_DEFAULT_FIRST_RESPONSE_LOW=24
SLA_DEFAULT_FIRST_RESPONSE_MEDIUM=8
SLA_DEFAULT_FIRST_RESPONSE_HIGH=4
SLA_DEFAULT_FIRST_RESPONSE_URGENT=2
SLA_DEFAULT_FIRST_RESPONSE_CRITICAL=1
SLA_DEFAULT_RESOLUTION_LOW=72
SLA_DEFAULT_RESOLUTION_MEDIUM=48
SLA_DEFAULT_RESOLUTION_HIGH=24
SLA_DEFAULT_RESOLUTION_URGENT=8
SLA_DEFAULT_RESOLUTION_CRITICAL=4

# Logging
LOG_LEVEL=info
LOG_ENCODING=json
LOG_DEVELOPMENT=false
LOG_FILE_PATH=logs/ticket.log
```

## Development

### Prerequisites
- Go 1.24+
- MongoDB 7+
- Redis 7+

### Running locally

```bash
# Download dependencies
make deps

# Run the service
make run

# Run with hot reload (requires air)
make dev
```

### Docker

```bash
# Start development environment
make docker-dev-up

# View logs
make docker-logs

# Stop containers
make docker-dev-down
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

## Architecture

```
ticket/
├── api/
│   ├── router/          # HTTP router setup
│   └── v1/handlers/     # HTTP handlers
├── cmd/
│   └── main.go          # Application entry point
├── config/              # Configuration
├── internal/
│   ├── database/        # Database connection
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Domain models
│   ├── repository/      # Data access layer
│   └── usecase/         # Business logic
├── locales/             # i18n translations
├── Dockerfile
├── docker-compose.yml
├── docker-compose.dev.yml
├── Makefile
└── README.md
```

## License

Copyright © Minisource
