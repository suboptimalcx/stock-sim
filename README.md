# Stock-Sim

A simplified, highly-available stock market simulator implemented in Go.

- The service is designed to be architecture portable and works on Linux, macOS, and Windows.
- High availability is supported through multiple app replicas behind an NGINX proxy.
- The implementation does not depend on wallet balances, only stock quantities and bank inventory.
- Following Go best practices, interfaces are defined in the packages that use them (the consumers) rather than the packages that implement them. This keeps the packages decoupled and prevents circular dependencies

This project exposes a REST API for:
- wallet stock operations (`buy` / `sell`)
- bank stock state management
- wallet state retrieval
- per-wallet stock quantities
- an audit log of successful wallet operations
- a chaos endpoint that terminates the current instance

The system is intentionally simplified:
- stock price is fixed at `1`
- wallets do not track balances or funds
- buy/sell operations are executed immediately
- the bank is the single liquidity provider
- the initial state has no wallets and an empty bank


---

## Features

- `POST /wallets/{wallet_id}/stocks/{stock_name}`
  - request body: `{ "type": "buy" | "sell" }`
  - creates the wallet if it does not exist
  - returns `404` when the requested stock does not exist in the bank
  - returns `400` when a buy is attempted with no bank stock or a sell with no wallet stock
  - returns `200` on success
- `GET /wallets/{wallet_id}`
  - returns wallet state including owned stocks and quantities
- `GET /wallets/{wallet_id}/stocks/{stock_name}`
  - returns wallet stock quantity as a plain number
- `GET /stocks`
  - returns bank stock state
- `POST /stocks`
  - request body: `{ "stocks": [{ "name": "stock1", "quantity": 99 }, ...] }`
  - sets the bank stock state
  - returns `200` on success
- `GET /log`
  - returns the ordered audit log of successful `buy` and `sell` operations
- `POST /chaos`
  - terminates the instance handling the request

---

## Architecture

To fulfill the requirements specifically high availability, concurrency control, and cross-platform simplicity, the following architecture was chosen:

*   **Language & Routing (Go + Chi):** Go was chosen for its excellent concurrency model, high performance, and standard library robustness. The `chi` router is used because it is lightweight, idiomatic, and relies entirely on standard `http.Handler`, keeping the footprint minimal.
*   **Database (PostgreSQL):** PostgreSQL was selected to ensure data integrity during concurrent `buy` and `sell` operations. (e.g., preventing race conditions where two simultaneous requests might successfully buy the last remaining stock).
*   **High Availability (NGINX + App Replicas):** To satisfy the requirement that killing an instance (`POST /chaos`) does not disrupt the product, the system runs **three parallel Go instances** behind an **NGINX reverse proxy/load balancer**. NGINX routes traffic using round-robin, ensuring seamless failover if a node drops.
*   **Layered Domain Design:** The codebase follows a standard service-repository pattern:
    *   `internal/handler`: HTTP request/response parsing and basic validation.
    *   `internal/service`: Core business logic (market rules, wallet validation).
    *   `internal/repository`: Data access layer and database transaction management.
*   **Infrastructure (Docker Compose):** Ensures the solution is entirely OS-agnostic (Windows, macOS, Linux, ARM64/x64) and starts with a single command without requiring the host machine to install anything but Docker.

### Project Structure
```text
.
├── cmd/api/             # Application entry point (main.go)
├── init/                # SQL initialization scripts for PostgreSQL
├── internal/
│   ├── handler/         # HTTP Layer & Routing
│   ├── service/         # Business Logic
│   ├── repository/      # PostgreSQL implementation & Transactions
│   └── model/           # Shared models & custom error types
├── Dockerfile           # Multi-stage build for the Go application
├── docker-compose.yml   # Infrastructure orchestration (App, DB, NGINX)
├── nginx.conf           # Load balancer configuration
├── start.sh / start.bat # Cross-platform startup scripts
└── go.mod               # Dependency management
```
---

## Prerequisites

- Docker
- Docker Compose
- Docker daemon running
- Go runtime for local test execution

---

## Running locally

### Linux / macOS

```bash
chmod +x start.sh
./start.sh
```

To use a custom port:

```bash
./start.sh 3000
```

### Windows

```bat
start.bat
```

To use a custom port:

```bat
start.bat 3000
```

If not specified, the service listens on `localhost:8080`.

### Stopping the service

```bash
docker compose down
```

To also remove the database volume:

```bash
docker compose down -v
```

---

## API Examples

Set bank stock state:

```bash
curl -X POST http://localhost:8080/stocks \
  -H "Content-Type: application/json" \
  -d '{"stocks":[{"name":"AAPL","quantity":100},{"name":"GOOG","quantity":50}]}'
```

Buy a stock for a wallet:

```bash
curl -X POST http://localhost:8080/wallets/user1/stocks/AAPL \
  -H "Content-Type: application/json" \
  -d '{"type":"buy"}'
```

Get wallet state:

```bash
curl http://localhost:8080/wallets/user1
```

Get a wallet stock quantity:

```bash
curl http://localhost:8080/wallets/user1/stocks/AAPL
```

Get bank stock state:

```bash
curl http://localhost:8080/stocks
```

Get audit log:

```bash
curl http://localhost:8080/log
```

Trigger chaos on the current instance:

```bash
curl -X POST http://localhost:8080/chaos
```

---

## Testing

### Unit tests
Thanks to the **Dependency Injection** pattern, the business logic in `internal/service` is tested in isolation. We inject mock implementations of the repository to simulate various database states and error conditions without the overhead of a real DB.
```bash
go test ./internal/service ./internal/handler -v
```

### Integration tests

These tests use the actual `PostgresRepository` to verify that the SQL queries and transaction logic (like `ON CONFLICT` and `FOR UPDATE`) work correctly against a real database schema.
Integration tests require a running PostgreSQL database. Start the database container first:

```bash
docker compose up db -d
```

Then run:

```bash
go test ./internal/repository -v
```

---

## Trade-offs & Notes

**Hardcoded Environment Variables:** Because this is an assessment task meant to be evaluated with zero configuration friction, database credentials are intentionally hardcoded in the docker-compose.yml. In a production environment, these would be managed via .env files or a secrets manager (like HashiCorp Vault or AWS Secrets Manager).

**Data Persistence:** The database utilizes Docker volumes to ensure data is not lost if the containers are restarted normally, though the start scripts may recreate the environment based on how they are configured.

**Static Services vs. Docker Replicas (NGINX DNS Caching):** Instead of utilizing Docker Compose's deploy: replicas: 3 feature, the application explicitly defines three distinct services (app1, app2, app3). This is a deliberate workaround for a known behavior in the open-source version of NGINX, which resolves upstream DNS names only once at startup. If a replica container were killed via the /chaos endpoint and recreated, Docker's internal DNS would assign it a new IP address, leaving NGINX routing traffic to a dead, cached IP. Defining distinct static services ensures more stable routing upon container restarts in this local environment. In a production environment requiring dynamic horizontal scaling, this would be addressed by using dynamic DNS resolvers in NGINX, adopting NGINX Plus, or relying on a dedicated orchestration platform like Kubernetes for service discovery.

---

## Author's Note
This project was developed for a backend internship application. Coming from an Embedded C and systems background, I used this task as an opportunity to learn Go and apply low-level principles like concurrency control and resource efficiency to a distributed environment.

I thoroughly enjoyed the shift to backend architecture and look forward to further applying my systems engineering mindset to scalable web services.
