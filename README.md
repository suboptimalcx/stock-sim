# stock-sim

A simple stock trading simulation API written in Go with PostgreSQL persistence and an Nginx load balancer.

## Endpoints

- `POST /wallets/{wallet_id}/stocks/{stock_name}`
  - Request body: `{ "type": "buy" }` or `{ "type": "sell" }`
  - Example: `POST /wallets/wallet123/stocks/ACME`
- `POST /chaos`
  - Terminates the app

## Run

1. Start the stack:

```bash
docker compose up --build
```