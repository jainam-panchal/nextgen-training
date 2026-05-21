# SSE PoC (Quick)

This PoC verifies live bid updates over Server-Sent Events.

## 1) Start server

```bash
go run ./cmd/server
```

Server runs on `http://localhost:8081`.

## 2) Seed users + item

```bash
# seller
curl -sS -X POST http://localhost:8081/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"seller","balance":100000}'

# bidder
curl -sS -X POST http://localhost:8081/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"bidder","balance":100000}'

# item (seller_id=1)
curl -sS -X POST http://localhost:8081/items \
  -H 'Content-Type: application/json' \
  -d '{"name":"iPhone 17","category_path":["Electronics","Phones"],"description":"demo","seller_id":1,"start_price":1000,"start_time":"2026-05-21T00:00:00Z","end_time":"2026-05-22T00:00:00Z"}'
```

## 3) Open SSE stream (terminal A)

```bash
curl -N http://localhost:8081/items/1/live
```

You should see heartbeat lines (`: ping`) and then event frames.

## 4) Place bid (terminal B)

```bash
curl -sS -X POST http://localhost:8081/items/1/bid \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: 2' \
  -d '{"user_id":2,"amount":1500}'
```

## 5) Expected SSE output

Terminal A should print something like:

```text
event: placed
data: {"item_id":1,"bid_id":1,"action":"placed","timestamp":"..."}
```
