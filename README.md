# Distributed In-Memory Counter with Service Discovery (Golang)

## Overview

This project implements a **distributed in-memory counter system** in Go where multiple nodes:

* Discover each other dynamically
* Maintain a consistent peer list
* Propagate counter updates
* Achieve **eventual consistency**
* Handle failures using retries and heartbeats

---

## System Design

### High-Level Architecture

```
        ┌──────────────┐
        │   Node A     │
        │ (Counter +   │
        │ Discovery)   │
        └──────┬───────┘
               │
     ┌─────────┼─────────┐
     │         │         │
┌────▼────┐ ┌──▼─────┐ ┌─▼──────┐
│ Node B  │ │ Node C │ │ Node D │
└─────────┘ └────────┘ └────────┘
```

Each node contains:

* **Counter Service** → manages increments
* **Discovery Service** → manages peers
* **Sync Service** → ensures eventual consistency
* **Registration Service** → handles join/leave

---

## 1. Service Discovery Design

### Node Registration

* Each node starts with a unique ID (`host:port`)
* Nodes are initialized with optional peers:

  ```
  --peers=localhost:8081,localhost:8082
  ```
* On startup:

  * Node registers itself
  * Syncs peer list with existing nodes

---

### Peer Synchronization

* Nodes exchange peer lists using:

  ```
  POST /sync-peers
  ```
* New peers are merged into local store
* Self and empty peers are ignored

---

### Heartbeat Mechanism

* Runs periodically (every 10s)
* Calls:

  ```
  GET /health
  ```
* If a node fails:

  * It is **removed from peer list**

---

### Peer Storage

* In-memory map:

  ```go
  map[string]bool
  ```
* Protected with `RWMutex` for concurrency safety

---

## 2. Distributed Counter Design

### Counter Operations

| Endpoint        | Description                |
| --------------- | -------------------------- |
| POST /increment | Increment counter          |
| GET /count      | Get current value          |
| POST /replicate | Apply replicated increment |

---

### Concurrency & Safety

* Uses `sync.Mutex` to avoid race conditions
* All updates are atomic

---

### Deduplication

Each increment has:

```go
ID = timestamp (int64)
```

* Stored in:

  ```go
  seenOps map[int64]bool
  ```
* Prevents duplicate increments across nodes

---

### Increment Flow

1. Client calls:

   ```
   POST /increment
   ```

2. Node:

   * Applies increment locally
   * Stores operation
   * Propagates to peers asynchronously

3. Peers:

   * Validate ID
   * Apply only if not seen before

---

### Propagation

* Uses async goroutines
* Sends request to:

  ```
  POST /replicate
  ```

---

## 3. Failure Handling & Retry

### Retry Strategy

* Uses exponential backoff:

  * Initial: 500ms
  * Max: 2s
  * Retries: 3

---

### Network Failure Handling

* If peer fails:

  * Retry with backoff
* If still failing:

  * Operation is eventually synced via SyncService

---

## 🔄 4. Sync Service (Eventual Consistency)

### Purpose:

Recover missed updates during failures.

### Flow:

1. Periodically fetch:

   ```
   GET /operations?since=timestamp
   ```

2. Apply missing operations

3. Update `lastSeen`

---

### Guarantees:

* Ensures **eventual consistency**
* Handles **network partitions**

---

## API Summary

### Counter APIs

```
POST /increment
GET  /count
POST /replicate
GET  /operations?since=<timestamp>
```

### Discovery APIs

```
GET  /peers
POST /sync-peers
GET  /heartbeat
```

### Registration

```
POST /register
```

---

##  Testing

* Framework: `testify`
* Mocking: `mockery`

### Covered Scenarios:

* Concurrent increments
* Deduplication
* Retry logic
* Peer failure detection
* Sync after partition

---

## How to Run

### 1. Start Node

```bash
go run main.go --port=8081 --peers=localhost:8082
```

### 2. Start Multiple Nodes

```bash
go run main.go --port=8082 --peers=localhost:8081
go run main.go --port=8083 --peers=localhost:8081
```

---

### 3. Test Increment

```bash
curl -X POST localhost:8081/increment
```

---

### 4. Get Count

```bash
curl localhost:8081/count
```

---

## 🌐 How System Handles Network Partitions

* Failed nodes are detected via heartbeat
* Missed updates are not lost:

  * Stored in operation log
  * Synced later via SyncService

This ensures:
**Eventual consistency across all nodes**

---

##  Limitations

1. In-memory only (no persistence)
2. No strong consistency guarantee (eventual only)
3. Timestamp-based IDs may collide in extreme cases
4. Network-heavy (full peer propagation)

---

## Design Decisions

### Why eventual consistency?

* Simpler
* Fits distributed counter use case

---

### Why deduplication via ID?

* Prevents double increments
* Enables safe retries

---

### Why pull-based sync?

* Recovers missed updates
* Handles node downtime

---

## 📂 Auther
Anand Gavai
