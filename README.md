# 🛒 Core Commerce / E-Commerce API

A scalable and reliable backend API for managing e-commerce orders, inventory, and secure authentication. Built with **Go**, **PostgreSQL**, and **Redis**, this project strongly focuses on preventing inventory overselling, ensuring transactional integrity during checkouts, and applying Clean Architecture principles.

## 🚀 Key Features

- **Overselling Prevention (Concurrency Control):** Mitigates database race conditions during flash sales or concurrent checkouts using row-level locking (`SELECT FOR UPDATE`). Ensures product inventory never drops below zero.
- **Atomic Order Processing:** Guarantees that the entire checkout flow—creating an order, recording order items, and deducting inventory—succeeds or fails together as a single ACID transaction.
- **High-Performance Caching & Secure Auth (Redis):** Utilizes **Redis** to cache frequently accessed product data, significantly reducing database load. Also implements secure session management by handling JWT token blocklists for the logout functionality.
- **Clean Architecture:** Strictly separates concerns across HTTP Handlers, Business Logic (Services), and Data Access (Repositories) to ensure maintainability, scalability, and high testability.
- **Fail-Fast Validation:** Validates core business rules directly at the Service layer (e.g., negative quantities, out-of-stock items, invalid SKUs) before initiating any database transactions, optimizing overall performance.

## 🗄️ Domain Models

The system focuses heavily on the transactional checkout phase and is built around three core entities:
1. **Product:** Stores essential catalog details (`sku`, `name`), pricing, real-time `stock`, and a `version` field for robust data state tracking.
2. **Order:** Records the overall purchase transaction, including the buyer's information, total price, and current fulfillment status.
3. **Order Item:** Captures a historical snapshot of the exact price and quantity of a product at the moment the order was placed (preventing past orders from changing if product prices update in the future).

## 🏗️ Project Structure & Initialization

> **💡 Note on Commit History:** This repository was initialized using a standard **[Go Starter Kit](https://github.com/codepnw/go-starter-kit)** to handle repetitive boilerplate code (e.g., basic folder structure, router setup). This strategic choice allowed the development focus to remain entirely on building the complex core business logic, database transactions, and concurrency handling that you will see in the subsequent commits.

The project directory follows a modular Go layout, combining **Clean Architecture** principles with a **Package by Feature** structure for better maintainability and high cohesion:


```text
.
├── cmd
│   └── api                 # Application entry point (main.go)
├── internal
│   ├── auth                # Context management & user session extraction
│   ├── config              # Environment variables & configuration loader
│   ├── errs                # Centralized custom error types
│   ├── features            # Domain modules (User, Product, Cart, Order) containing their own Handler, Service, and Repo
│   ├── middleware          # HTTP middlewares (e.g., JWT Auth, Logger, Recovery)
│   └── server              # HTTP server initialization & route registration
├── pkg
│   ├── database            # Database connection & transaction management
│   ├── jwttoken            # JWT generation & validation utilities
│   └── utils               # Shared helper functions (e.g., password hashing)
├── .air.toml               # Live reload configuration for rapid local development
├── .env.example            # Template for environment variables
├── docker-compose.yml      # Local development environment setup (PostgreSQL, etc.)
├── Dockerfile              # Docker build instructions for production deployment
└── Makefile                # Shortcut commands for build, test, migrate, and run
```

## 🚀 Getting Started

Follow these steps to get the project up and running on your local machine.

### Option 1: Quick Start with Docker (Recommended) 🐳

This will spin up both the Go API server and the PostgreSQL database container.

1.  **Clone the repository**
    ```bash
    git clone https://github.com/codepnw/go-core-ecommerce.git

    cd go-core-ecommerce
    ```

2.  **Setup Environment Variables**
    ```bash
    cp -n .env.example .env
    ```
    *Modify `.env` if you want to change default ports or secrets.*

3.  **Start Services**
    ```bash
    # Build and start both App, Redis & DB containers
    docker compose up -d --build
    ```
    *(Note: If you only want to run the database container, use `docker compose up -d db`)*

4.  **Run Database Migrations**
    ```bash
    # If you have Makefile configured
    make migrate-up
    
    # Or manually using golang-migrate
    migrate -path [MIGRATION_PATH] -database [DATABASE_URL] up
    ```

The API will be available at `http://localhost:8080/api/v1` (Default URL).

### Option 2: Run Locally (Without Docker)

If you prefer to run the Go application directly on your host machine:

1.  **Start PostgreSQL** (Make sure you have a running instance).
2.  **Update `.env`** to point to your local DB credentials.
3.  **Run the application**:
    ```bash
    # If you have Makefile configured
    make run

    # If you have Air Live Reload
    air

    # Or run the standard Go command
    go run cmd/api/main.go

    ```

---

## 📡 API Endpoints Summary

The core business logic of this E-commerce API focuses on cart management and order checkout processing. Endpoints marked with 🔒 require an `Authorization: Bearer <token>` header.

*(Base URL: `http://localhost:8080/api/v1`)*

### 🛒 Cart
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :---: |
| `GET` | `/cart` | Get current user's cart | 🔒 |
| `POST` | `/cart/items` | Add item to cart | 🔒 |
| `DELETE` | `/cart/items/:product_id` | Remove item from cart | 🔒 |

### 🧾 Orders (Checkout & Transactions)
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :---: |
| `GET` | `/orders` | Get current user's order history | 🔒 |
| `POST` | `/orders/checkout` | Checkout / Create a new order | 🔒 |
| `GET` | `/orders/:order_id` | Get order details by ID | 🔒 |
| `PATCH` | `/orders/:order_id/status` | Update order status | 👑 **Admin** |

### 📦 Products
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :---: |
| `GET` | `/products` | Get all products | ❌ |
| `GET` | `/products/:product_id` | Get product details by ID | ❌ |
| `POST` | `/products` | Create a new product | 🔒 |
| `PATCH` | `/products/:product_id` | Update product details | 🔒 |
| `DELETE` | `/products/:product_id` | Delete a product | 🔒 |
| `POST` | `/products/:product_id/stock`| Increase product stock | 🔒 |

### 👤 Auth & Users
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :---: |
| `POST` | `/auth/register` | Register a new user | ❌ |
| `POST` | `/auth/login` | User login | ❌ |
| `POST` | `/auth/refresh-token` | Refresh access token | ❌ |
| `POST` | `/auth/logout` | User logout | 🔒 |
| `GET` | `/users/profile` | Get current user profile | 🔒 |

### 🟢 Health Check
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :---: |
| `GET` | `/health` | Check API health status | ❌ |

---