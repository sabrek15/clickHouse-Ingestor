# ClickHouse Ingestor

![Go](https://img.shields.io/badge/Go-1.20+-blue)
![ClickHouse](https://img.shields.io/badge/ClickHouse-22.8+-yellow)

A lightweight web app for bidirectional data transfer between ClickHouse and CSV files.

## 🚀 Quick Start

### Prerequisites
- [Go 1.20+](https://go.dev/dl/)
- [ClickHouse server](https://clickhouse.com/docs/en/install) (local or remote)

### 1. Clone & Install
```bash
git clone https://github.com/sabrek15/clickHouse-Ingestor.git
cd clickHouse-Ingestor
go mod download
```

## ⚙️ Configuration

1. **Environment Variables**  
   Create a `.env` file in the root directory with the following variables:
   ```env
   JWT_SECRET="your-jwt-secret"
   CLICKHOUSE_JWT_SECRET="your-clickhouse-jwt-secret"
   PORT=8080
   ```

   - `JWT_SECRET`: Secret key for generating and validating JWT tokens.
   - `CLICKHOUSE_JWT_SECRET`: Secret key for ClickHouse JWT authentication.
   - `PORT`: Port on which the web app will run (default is `8080`).

2. **ClickHouse Server**  
   Ensure you have a running ClickHouse server. You can install it locally or use a remote instance. Refer to the [ClickHouse installation guide](https://clickhouse.com/docs/en/install).

## ▶️ Running the Web App

1. **Start the Server**
   Run the following command to start the web app:
   ```bash
   go run main.go
   ```

2. **Access the Web App**
   Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

## 🛠️ Features

- **ClickHouse Integration**: Connect to a ClickHouse server and perform schema discovery, data export, and import.
- **CSV File Handling**: Read and write CSV files with customizable delimiters.
- **JWT Authentication**: Secure API endpoints with JWT-based authentication.
- **Web UI**: Intuitive interface for configuring data sources and initiating transfers.

## 📂 Project Structure

```plaintext
clickHouse-Ingestor/
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── README.md
├── internal/
│   ├── api/
│   │   └── auth.go
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── middleware.go
│   │   └── password.go
│   ├── filehandler/
│   │   └── csv.go
│   └── storage/
│       ├── clickhouse.go
│       └── clickhouse_test.go
├── static/
│   ├── index.html
│   ├── css/
│   │   ├── auth.css
│   │   ├── styles.css
│   │   └── tables.css
│   └── js/
│       └── app.js
```

## 📖 API Endpoints

### `/api/connect`  
**Method**: `POST`  
**Description**: Connect to a ClickHouse server.  
**Request Body**:
```json
{
  "host": "localhost",
  "port": "8123",
  "database": "default",
  "user": "default",
  "jwtToken": "your-jwt-token",
  "secure": false
}
```

### `/api/discover-schema`  
**Method**: `GET`  
**Description**: Discover schema from ClickHouse or a CSV file.  
**Query Parameters**:
- `source`: `clickhouse` or `file`
- `filePath` (for `file` source): Path to the CSV file.
- `delimiter` (optional): Delimiter used in the CSV file.

### `/api/transfer`  
**Method**: `POST`  
**Description**: Transfer data between ClickHouse and CSV files.  
**Request Body**:
```json
{
  "sourceType": "clickhouse",
  "sourceParams": { ... },
  "targetType": "file",
  "targetParams": { ... },
  "columns": ["column1", "column2"],
  "table": "table_name"
}
```

## 🧪 Running Tests

Run the following command to execute tests:
```bash
go test ./internal/storage
```
