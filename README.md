# Polyglot Portfolio - Go API

[![Go Version](https://img.shields.io/badge/Go-v1.25.x-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Docker Support](https://img.shields.io/badge/Docker-Support-2496ED?logo=docker&logoColor=white)](https://www.docker.com/) [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A high-performance backend REST API built with Go (Golang) to serve content for the modern personal portfolio frontend. This is one of the three backend implementations demonstrating cross-stack compatibility.

---

## Key Features

*   **Authentication**: Fast and secure JWT-based login for the admin dashboard.
*   **Content Management API**: High-throughput CRUD endpoints for managing projects, work experiences, technology stacks, and categories.
*   **Static Asset Serving**: Handling image and asset uploads efficiently.

---

## Tech Stack & Libraries

*   **Language & Version**: Go (v1.25.5)
*   **Web Framework**: Gin Gonic (v1.11.0)
*   **Database / Driver**: PostgreSQL (via `pgx/v5` driver)
*   **Caching**: Redis (via `go-redis/v9`)
*   **Emailing**: Resend Go SDK (v2)
*   **Authentication**: JWT (`golang-jwt/jwt/v5`) & Bcrypt (`golang.org/x/crypto`)
*   **Rate Limiting**: Ulule Limiter (`ulule/limiter/v3`)

---

## Installation & Setup

### Method 1: Manual Local Setup

#### Prerequisites
*   Go >= 1.20
*   A running database instance (e.g., MySQL or PostgreSQL)

#### Steps
1.  **Clone and Download Dependencies**:
    ```bash
    git clone <repository-url>
    cd go-portfolio-api
    go mod download
    ```
2.  **Environment Setup**:
    Copy `.env.example` to `.env` and fill in DB credentials and JWT secrets.
3.  **Run Application**:
    ```bash
    go run main.go
    ```

---

## Author

**Churma16**

---

## License

This project is licensed under the [MIT License](LICENSE).
