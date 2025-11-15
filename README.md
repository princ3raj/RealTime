# Go Real-Time Messaging Service

A high-performance backend for a real-time messaging application (e.g., WhatsApp-like) written in Go. This project is built using a production-grade, scalable, and maintainable **Clean Architecture** pattern.

It features a high-performance **WebSocket server** for real-time communication and a separate **REST API server** for authentication and supporting operations.

## 🏗️ Architectural Overview

This project is not a simple monolith; it's a professionally structured application designed for separation of concerns, testability, and scalability.

The architecture is built on **Domain-Driven Design (DDD)** principles:

* **`cmd/` (Entry Points):** Contains the `main.go` files for our two distinct server binaries:
    * `cmd/api`: The REST API server (handles login, registration, etc.).
    * `cmd/server`: The WebSocket server (handles all real-time connections).
* **`internal/api/` (Handler Layer):** The "Controller" layer. It knows how to speak HTTP and WebSocket. Its only job is to parse requests, call services, and return responses.
    * `/rest`: Handlers for the REST API.
    * `/ws`: Handlers for upgrading WebSocket connections.
* **`internal/domain/` (Business Logic Layer):** The heart of the application. Contains all business logic, models, and services. This layer has **zero** knowledge of the web or the database.
* **`internal/store/` (Persistence Layer):** The database (repository) layer. Contains all SQL queries and logic for persisting data.
* **`internal/app/` (WS Concurrency):** A specialized package containing the core concurrency model (Hub/Client) for managing thousands of WebSocket connections efficiently.
* **`internal/wsrouter/` (WS Routing):** A dedicated router that inspects incoming WebSocket message `types` (e.g., "chat", "typing") and routes them to the correct handler, preventing a monolithic Hub.

## ✨ Key Features

* **Dual-Server Architecture:** Separates HTTP request/response logic (`cmd/api`) from persistent connection logic (`cmd/server`).
* **High-Performance Concurrency:** Utilizes Go's standard library for a highly concurrent Hub/Client pattern to manage tens of thousands of connections.
* **Clean Architecture:** Strict separation of concerns between API handlers, domain logic, and data storage.
* **Structured Logging:** Configured with `go.uber.org/zap` for high-performance, structured, production-ready logging.
* **Secure Configuration:** Uses `viper` to load configuration and secrets (like `JWT_SECRET`) from environment variables, never code.
* **JWT Authentication:** All endpoints are secured using JWT, including the WebSocket upgrade handshake.
* **Message Routing:** `wsrouter` cleanly handles different message types (`chat`, `join`, `private`) without bloating the connection Hub.

---

##  Directory Structure

`````├── cmd
│   ├── api
│   │   └── main.go
│   └── server
│       └── main.go
├── go.mod
├── go.sum
├── internal
│   ├── api
│   │   ├── rest
│   │   │   └── handler.go
│   │   └── ws
│   │       └── handler.go
│   ├── app
│   │   ├── client.go
│   │   ├── hub.go
│   │   └── message.go
│   ├── auth
│   │   └── auth.go
│   ├── config
│   │   └── config.go
│   ├── domain
│   ├── log
│   │   └── log.go
│   ├── store
│   └── wsrouter
│       ├── handlers.go
│       └── router.go
├── pkg
│   └── websocketutil
└── ws_client.html
``````
## 🚀 Getting Started

### Prerequisites

* Go 1.21+
* PostgreSQL (or a compatible database)
* Access to a terminal

### 1. Installation

1.  **Clone the repository:**
    ```sh
    git clone https://your-repo-url/RealTime.git
    cd RealTime
    ```
2.  **Install dependencies:**
    ```sh
    go mod tidy
    ```

### 2. Configuration

This application **requires** environment variables to run. Create a `.env` file (or `export` these in your shell) with the following values:

```sh
# .env

# Critical secret for signing JWT tokens. Must be identical for both servers.
# Use a strong, randomly generated string.
export JWT_SECRET='your_super_strong_random_secret_key'

# PostgreSQL connection string
export DB_URL='user=postgres password=mysecretpassword dbname=mydb sslmode=disable'

# Ports for the two servers
export API_PORT='8081'      # Port for the REST API server
export SERVER_PORT='8080'   # Port for the WebSocket server