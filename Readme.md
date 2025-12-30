# LevPay Backend 🚀

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Flavor](https://img.shields.io/badge/flavor-fiber-green?style=for-the-badge)
![Database](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)

**LevPay Backend** is the high-performance core of the LevPay financial platform. Built with **Go** and **Fiber**, it provides a robust, scalable, and secure API for managing digital wallets, transactions, users, and financial operations.

## ✨ Key Features

-   **Clean Architecture**: Modular and testable code structure.
-   **Security First**: JWT Authentication, Role-Based Access Control (RBAC), and bcrypt hashing.
-   **Financial Engine**: Transaction management (Deposits, Transfers, Withers), Wallet logic, and Invoice billing.
-   **Real-time Notifications**: Integrated notification system.
-   **KYC Verification**: Multi-step identity verification service with document handling (MinIO).
-   **Admin Dashboard**: Comprehensive admin endpoints for system monitoring and user management.
-   **Documentation**: Auto-generated Swagger UI.

## 🛠️ Tech Stack

-   **Language**: Go 1.22+
-   **Framework**: [Fiber](https://gofiber.io/) (Fastest Go Web Framework)
-   **Database**: PostgreSQL
-   **ORM**: GORM
-   **Storage**: MinIO (S3 Compatible)
-   **Message Queue**: RabbitMQ (for event-driven architecture)
-   **Environment**: Docker & Docker Compose

## 🚀 Getting Started

### Prerequisites

-   Docker & Docker Compose
-   Go 1.22+ (optional, for local dev without Docker)

### Quick Start (Recommended)

We provide a luxury management script to handle everything for you.

1.  **Start Services**:
    ```bash
    ./manage.sh up
    ```
    This will spin up Postgres, MinIO, RabbitMQ, and the Backend API.

2.  **Check Health**:
    ```bash
    ./manage.sh health
    ```

3.  **View Logs**:
    ```bash
    ./manage.sh logs
    ```

4.  **Stop Services**:
    ```bash
    ./manage.sh down
    ```

### 📚 API Documentation

Once the server is running, explore the full API reference via Swagger UI:

👉 **[http://localhost:5001/swagger/index.html](http://localhost:5001/swagger/index.html)**

## 📂 Project Structure

```
levpay-backend/
├── cmd/                # Entry points for services (app, admin, cron, etc.)
├── feature/            # Feature-based logic (User, Wallet, Transaction, etc.)
│   ├── [feature]/
│   │   ├── handler.go    # HTTP Controllers
│   │   ├── repository.go # Database Data Access
│   │   └── service.go    # Business Logic
├── internal/           # Private application code
│   ├── config/         # Configuration loader
│   ├── database/       # DB Connection setup
│   ├── middleware/     # Auth & Logging middleware
│   └── models/         # Database structs
├── router/             # API Route definitions
├── docs/               # Swagger Documentation
├── docker-compose.yaml # Production Orchestration
└── manage.sh           # Management Utility
```

## 🔐 Environment Variables

Copy `.env.example` to `.env` and configure your credentials:

```bash
cp .env.example .env
```

Key variables:
-   `DB_HOST`, `DB_USER`, `DB_PASSWORD`: Database connection
-   `JWT_SECRET`: Secret key for token signing
-   `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`: Object storage config

## 🤝 Contributing

1.  Fork the repository
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request
