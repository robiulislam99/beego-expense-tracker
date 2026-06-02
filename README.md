# Expense Tracker API

A RESTful API for tracking personal expenses, built with **Go** and **Beego v2**. All data is stored in CSV files using Go's standard `encoding/csv` library.

---

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Beego v2
- **Storage:** CSV files (standard library)
- **Documentation:** Swagger UI (swaggo)

---

## Prerequisites

Make sure the following tools are installed before running the project.

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.22+ | https://go.dev/dl |
| Bee CLI | v2 | `go install github.com/beego/bee/v2@latest` |
| Git | Any | https://git-scm.com |

---

## Setup

### 1. Clone the repository

```bash
git clone https://github.com/robiulislam99/beego-expense-tracker
cd beego-expense-tracker
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Configure the app

```bash
cp conf/app.conf.sample conf/app.conf
```

The default `app.conf` values work out of the box — no changes needed to run locally.

### 4. Run the server

```bash
bee run
```

Server starts at: `http://localhost:8080`

The `data/` folder and CSV files are created automatically on first request.

---

## Swagger Documentation

Once the server is running, open the Swagger UI in your browser:

```
http://localhost:8080/swagger/index.html
```

All endpoints are documented with request parameters, body schemas, and response examples.

---

## API Endpoints

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| GET | `/api/v1/health` | No | Health check |
| POST | `/api/v1/auth/register` | No | Register a new user |
| POST | `/api/v1/auth/login` | No | Login and get user ID |
| POST | `/api/v1/expenses` | Yes | Create a new expense |
| GET | `/api/v1/expenses` | Yes | List expenses (with filters) |
| GET | `/api/v1/expenses/summary` | Yes | Spending summary by category |
| GET | `/api/v1/expenses/:id` | Yes | Get a single expense |
| PUT | `/api/v1/expenses/:id` | Yes | Update an expense |
| DELETE | `/api/v1/expenses/:id` | Yes | Delete an expense |

> **Authentication:** After login, pass the returned `user_id` as the `X-User-ID` header on all expense requests.

---

## Allowed Expense Categories

```
Food | Transport | Housing | Entertainment | Shopping
Healthcare | Education | Utilities | Other
```

---

## Sample curl Commands

### Health Check

```bash
curl -X GET http://localhost:8080/api/v1/health
```

**Response:**
```json
{
  "success": true,
  "message": "Server is running"
}
```

---

### Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "secret123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully"
}
```

---

### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "secret123"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user_id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

> Copy the `user_id` from the response and use it as the `X-User-ID` header for all expense requests.

---

### Create an Expense

```bash
curl -X POST http://localhost:8080/api/v1/expenses \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "title": "Lunch",
    "amount": 350.50,
    "category": "Food",
    "note": "Team lunch",
    "expense_date": "2025-06-10"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Expense created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "Lunch",
    "amount": 350.5,
    "category": "Food",
    "note": "Team lunch",
    "expense_date": "2025-06-10",
    "created_at": "2025-06-10T14:30:00Z"
  }
}
```

---

### List All Expenses

```bash
curl -X GET http://localhost:8080/api/v1/expenses \
  -H "X-User-ID: 1"
```

---

### List Expenses with Filters

Filter by category:
```bash
curl -X GET "http://localhost:8080/api/v1/expenses?category=Food" \
  -H "X-User-ID: 1"
```

Filter by date range:
```bash
curl -X GET "http://localhost:8080/api/v1/expenses?date_from=2025-06-01&date_to=2025-06-30" \
  -H "X-User-ID: 1"
```

Sort by amount descending:
```bash
curl -X GET "http://localhost:8080/api/v1/expenses?sort_by=amount&sort_order=desc" \
  -H "X-User-ID: 1"
```

Sort by date ascending with limit:
```bash
curl -X GET "http://localhost:8080/api/v1/expenses?sort_by=expense_date&sort_order=asc&limit=5" \
  -H "X-User-ID: 1"
```

Combined filters:
```bash
curl -X GET "http://localhost:8080/api/v1/expenses?category=Food&date_from=2025-06-01&sort_by=amount&sort_order=desc" \
  -H "X-User-ID: 1"
```

---

### Get a Single Expense

```bash
curl -X GET http://localhost:8080/api/v1/expenses/1 \
  -H "X-User-ID: 1"
```

---

### Update an Expense

```bash
curl -X PUT http://localhost:8080/api/v1/expenses/1 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 1" \
  -d '{
    "title": "Team Dinner",
    "amount": 500.00,
    "category": "Food",
    "note": "Updated note",
    "expense_date": "2025-06-10"
  }'
```

---

### Delete an Expense

```bash
curl -X DELETE http://localhost:8080/api/v1/expenses/1 \
  -H "X-User-ID: 1"
```

---

### Spending Summary

```bash
curl -X GET "http://localhost:8080/api/v1/expenses/summary?date_from=2025-06-01&date_to=2025-06-30" \
  -H "X-User-ID: 1"
```

**Response:**
```json
{
  "success": true,
  "message": "Summary generated",
  "data": {
    "date_from": "2025-06-01",
    "date_to": "2025-06-30",
    "total_amount": 15230.50,
    "total_count": 23,
    "by_category": [
      { "category": "Food", "total": 5400.00, "count": 12 },
      { "category": "Transport", "total": 3200.00, "count": 8 }
    ]
  }
}
```

---

## Running Tests

```bash
go test -v -coverpkg=./... ./tests/...
```

Current coverage: **73.5%**

---

## Query Parameters Reference

| Parameter | Type | Endpoint | Description |
|-----------|------|----------|-------------|
| `category` | string | GET /expenses | Filter by category name |
| `date_from` | YYYY-MM-DD | GET /expenses | Filter expenses on or after this date |
| `date_to` | YYYY-MM-DD | GET /expenses | Filter expenses on or before this date |
| `sort_by` | string | GET /expenses | Sort by `amount` or `expense_date` |
| `sort_order` | string | GET /expenses | `asc` or `desc` (default: `desc`) |
| `limit` | integer | GET /expenses | Maximum number of results to return |
| `date_from` | YYYY-MM-DD | GET /expenses/summary | Required — summary start date |
| `date_to` | YYYY-MM-DD | GET /expenses/summary | Required — summary end date |

---

## Validation Rules

**Register:**
- `name` — required
- `email` — required, must be valid email format
- `password` — required, minimum 6 characters
- `email` must be unique

**Expense:**
- `title` — required
- `amount` — required, must be a positive number
- `category` — required, must be one of the allowed categories
- `expense_date` — required, must be in `YYYY-MM-DD` format
- `note` — optional

---

## Project Structure

```
beego-expense-tracker/
├── conf/
│   ├── app.conf.sample       # Sample config (commit this)
│   └── app.conf              # Real config (git ignored)
├── controllers/
│   ├── auth.go               # Register, Login, HealthCheck
│   ├── base.go               # Shared response helpers
│   └── expense.go            # Expense CRUD, filters, summary
├── models/
│   ├── user.go               # User struct + CSV functions
│   └── expense.go            # Expense struct + CSV functions
├── routers/
│   └── router.go             # Route registration
├── docs/                     # Auto-generated Swagger docs
├── swagger/                  # Swagger UI static files
├── data/                     # CSV files (auto-created, git ignored)
├── tests/
│   ├── auth_test.go          # User model tests
│   └── expense_test.go       # Expense model tests
├── main.go
├── go.mod
├── go.sum
└── README.md
```
