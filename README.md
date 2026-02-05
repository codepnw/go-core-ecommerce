# Go Clean Architecture Starter Kit 🚀


A production-ready boilerplate for building RESTful APIs in Go, featuring **Clean Architecture**, **Dependency Injection**, and **Docker** support. Designed to be scalable, testable, and maintainable.

## 🌟 Key Features

* **Clean Architecture**: Separation of concerns (Handler -> Service -> Repository).
* **Gin Framework**: Fast and lightweight HTTP web framework.
* **PostgreSQL**: Robust relational database with **Migrations**.
* **JWT Authentication**: Secure Access Token & Refresh Token mechanism.
* **Dependency Injection**: Manually wired for better control and testing without reflection magic.
* **Docker Support**: Ready-to-ship with `docker-compose`.
* **Configuration**: Environment variable management using  `godotenv`.
* **Graceful Shutdown**: Handles OS signals to close connections properly.

## 🛠 Tech Stack

* **Language**: Go (Golang)
* **Framework**: Gin
* **Database**: PostgreSQL
* **Migration**: Golang-Migrate (or your preferred tool)
* **Auth**: JWT (JSON Web Tokens)
* **Containerization**: Docker & Docker Compose

## 📂 Project Structure

This project adopts the **Standard Go Project Layout**, organized by **features** to maintain modularity and Clean Architecture principles:

```text
├── cmd
│   └── api             # Application entry point (main.go)
├── internal            # Private application code (not importable by other projects)
│   ├── auth            # Authentication logic & context
│   ├── config          # Configuration loader (Environment variables)
│   ├── errs            # Custom error definitions and codes
│   ├── features        # Feature modules (User, Product, etc.) containing Handler, Service, Repo
│   ├── middleware      # HTTP Middlewares (Auth, CORS, Logger)
│   └── server          # Server initialization and graceful shutdown logic
├── pkg                 # Public shared libraries
│   ├── database        # Database connection setup & Migration helpers
│   ├── jwttoken        # JWT token generation and parsing utilities
│   └── utils           # Common utilities (Password hashing, Response format, Validation)
├── .env.example        # Example environment variables
├── docker-compose.yml  # Local development environment setup
├── Dockerfile          # Docker build instructions
└── Makefile            # Make commands for build, test, and run
```

## 🚀 Getting Started

Follow these steps to get the project up and running on your local machine.

### Prerequisites

* [Go 1.24+](https://go.dev/dl/)
* [Docker](https://www.docker.com/) & Docker Compose
* [Make](https://www.gnu.org/software/make/) (Optional, for using Makefile commands)

### Option 1: Quick Start with Docker (Recommended) 🐳

This will spin up both the Go API server and the PostgreSQL database container.

1.  **Clone the repository**
    ```bash
    git clone https://github.com/codepnw/go-starter-kit.git
    cd go-starter-kit
    ```

2.  **Setup Environment Variables**
    ```bash
    cp .env.example .env
    ```
    *Modify `.env` if you want to change default ports or secrets.*

3.  **Start Services**
    ```bash
    docker compose up -d --build
    ```

4.  **Run Database Migrations**
    ```bash
    # If you have Makefile configured
    make migrate-up
    
    # Or manually using golang-migrate
    migrate -path [MIGRATION_PATH] -database [DATABASE_URL] up
    ```

The API will be available at `http://localhost:8080`.

### Option 2: Run Locally (Without Docker)

If you prefer to run the Go application directly on your host machine:

1.  **Start PostgreSQL** (Make sure you have a running instance).
2.  **Update `.env`** to point to your local DB credentials.
3.  **Run the application**:
    ```bash
    go run cmd/api/main.go
    ```

---

## 🔌 API Endpoints

This starter kit comes with a fully functional Authentication feature.

### 🔐 Authentication (`/api/v1/auth`)

| Method | Endpoint | Description | Auth Header |
| :--- | :--- | :--- | :--- |
| `POST` | `/register` | Register a new user account | ❌ |
| `POST` | `/login` | Login to receive Access & Refresh Tokens | ❌ |
| `POST` | `/refresh` | Exchange Refresh Token for a new Access Token | ❌ |
| `POST` | `/logout` | Revoke the current Refresh Token | ✅ |

### 👤 User Profile (`/api/v1/users`)

| Method | Endpoint | Description | Auth Header |
| :--- | :--- | :--- | :--- |
| `GET` | `/profile` | Get the currently logged-in user's profile | ✅ `Bearer <token>` |

---

## 🔧 Configuration

The application is configured via environment variables. Copy `.env.example` to `.env` to customize.

```ini
# ⚠️ Default Base URL: http://localhost:8080/api/v1 ⚠️
# APP_HOST=localhost
# APP_PORT=8080
# APP_PREFIX=/api/v1

# -------------------------------------------------
# 🐘 DATABASE (PostgreSQL)
# ⚠️ Warning: Must Change in Production ⚠️
# -------------------------------------------------
DB_USER=postgres                
DB_PASSWORD=go-starter-kit-password            
DB_NAME=gostarter                
DB_HOST=localhost
# DB_PORT Matches docker-compose mapping
DB_PORT=5433
DB_SSL_MODE=disable

# ---------------------------------------
# 🔐 JWT SECURITY
# ⚠️ Warning: Must Change in Production ⚠️
# ---------------------------------------
JWT_APP_NAME=go-starter-kit_Change-in-Production           
JWT_SECRET_KEY=go-starter-kit_secret-key_Change-in-Production  
JWT_REFRESH_KEY=go-starter-kit-refresh-key_Change-in-Production

