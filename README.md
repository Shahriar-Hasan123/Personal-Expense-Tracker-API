# Personal Expense Tracker API

A RESTful API for managing personal expenses with user authentication and comprehensive expense tracking features.

## Overview

The Personal Expense Tracker API is a production-ready Go application built with the Beego framework. It provides endpoints for user registration, authentication, and complete expense management with filtering, sorting, and summary capabilities.

## Features

- **User Management**: Register and authenticate users with secure password handling
- **Expense Operations**: Create, read, update, and delete expenses
- **Advanced Filtering**: Filter expenses by category and date range
- **Sorting & Pagination**: Sort by amount or date with customizable limits
- **Expense Summary**: Generate summaries with per-category breakdowns
- **Health Check**: Monitor API availability
- **API Documentation**: Interactive Swagger UI with complete endpoint documentation

## Tech Stack

- **Language**: Go 1.26
- **Framework**: Beego v2.1.0
- **Documentation**: Swagger/OpenAPI 2.0
- **Testing**: Go's built-in testing package
- **Data Storage**: CSV files

## Prerequisites

- Go 1.26 or later
- Git
- Bee CLI (optional, for hot-reloading during development)

## Installation

### 1. Clone the repository
```bash
git clone https://github.com/Shahriar-Hasan123/Personal-Expense-Tracker-API.git
cd Personal-Expense-Tracker-API
```

### 1.1 Create local config
After cloning, create `conf/app.conf` in the `conf` folder by copying the example configuration file:
```bash
cp conf/app.conf.example conf/app.conf
```

### 2. Install dependencies
```bash
go mod download
go mod tidy
```

### 3. Install Bee CLI (optional)
```bash
go install github.com/beego/bee/v2@latest
```

### 4. Install Swagger tools (optional, only if regenerating docs)
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Running the Application

### Development mode (with hot-reload)
```bash
bee run
```
The API will start at `http://localhost:8080`

### Production mode
```bash
go build -o expense-tracker-api
./expense-tracker-api
```

## API Documentation

Once the server is running, access the interactive Swagger UI at:
```
http://localhost:8080/swagger/
```

### API Base URL
```
http://localhost:8080/api/v1
```

### Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Authenticate user |
| GET | `/expenses` | List all expenses (with filters) |
| POST | `/expenses` | Create a new expense |
| GET | `/expenses/{id}` | Get a single expense |
| PUT | `/expenses/{id}` | Update an expense |
| DELETE | `/expenses/{id}` | Delete an expense |
| GET | `/expenses/summary` | Get expense summary by category |
| GET | `/health` | Check API health status |

## Project Structure

```
.
├── conf/                 # Configuration files
├── controllers/          # HTTP request handlers
├── data/                 # CSV data storage
│   ├── users.csv
│   └── expenses.csv
├── docs/                 # Auto-generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── models/               # Data models and business logic
├── routers/              # Route definitions
├── swagger/              # Swagger UI static files
├── main.go               # Application entry point
├── go.mod                # Go module definition
└── go.sum                # Go module checksums
```

## Testing

Run the complete test suite with coverage:
```bash
go test ./... -v -cover
```
Run the complete test suite with coverage (without verbose output):
```bash
go test ./... -cover
```

Run tests for a specific package:
```bash
go test ./controllers -v
go test ./models -v
```

Generate coverage report:
```bash
go test ./... -coverprofile=coverage.out
```

## Usage Examples

### Register a new user
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

### Create an expense
```bash
curl -X POST http://localhost:8080/api/v1/expenses \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "title": "Groceries",
    "amount": 45.50,
    "category": "food",
    "note": "Weekly shopping",
    "expense_date": "2026-06-02"
  }'
```

### List expenses with filters
```bash
curl "http://localhost:8080/api/v1/expenses?category=food&sort_by=amount&sort_order=desc" \
  -H "X-User-ID: 1"
```

### Get expense summary
```bash
curl "http://localhost:8080/api/v1/expenses/summary?date_from=2026-06-01&date_to=2026-06-30" \
  -H "X-User-ID: 1"
```

## Authentication

The API uses header-based authentication via the `X-User-ID` header. Include this header with every request that requires authentication:

```
X-User-ID: <user_id>
```

## Configuration

Application configuration is loaded from `conf/app.conf`.

## Development

### Generate/Update Swagger Documentation
After modifying API endpoints, regenerate the Swagger spec:
```bash
swag init
```

### Project Standards
- Follow Go code conventions and best practices
- Current test coverage:
  - `models`: 87.6% (excellent)
  - `controllers`: 79.5% (good)
- To calculate overall coverage, run:
  ```bash
  go test ./... -coverprofile=coverage.out
  go tool cover -func=coverage.out | grep total
  ```
- Use meaningful commit messages following conventional commits
- Document all exported functions and types

## Known Issues & Notes

- The application uses CSV files for data storage, suitable for development/learning
- In production, replace CSV storage with a proper database (PostgreSQL, MySQL, etc.)
- Beego config warning about missing `conf/app.conf` is harmless if not required

## Future Improvements

- [ ] Switch from CSV to relational database (PostgreSQL)
- [ ] Add JWT token-based authentication
- [ ] Implement user authorization levels
- [ ] Add expense categories management
- [ ] Implement recurring expenses
- [ ] Add data export (PDF, Excel)
- [ ] Mobile app support

**Last Updated**: June 2, 2026  
**API Version**: 1.0  
**Go Version**: 1.26+
