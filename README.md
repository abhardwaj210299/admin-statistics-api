# Casino Stats API

A lightweight API for analyzing casino game transactions built with Go, MongoDB, and Redis.

## Quick Setup

### Prerequisites

- Go 1.19+
- Docker (for MongoDB and Redis)

### 1. Clone and Install

```bash
# Get the code
git clone https://github.com/abhardwaj210299/admin-statistics-api.git
cd admin-statistics-api

# Install dependencies
go mod download
```

### 2. Start Databases

```bash
# Start MongoDB
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Start Redis
docker run -d -p 6379:6379 --name redis redis:latest
```

### 3. Generate Data

```bash
# Create sample transactions (2M rounds)
go run cmd/seed/data-seed.go
```

### 4. Run the API

```bash
# Start the server
go run cmd/api/api-main.go
```

The API runs at http://localhost:8080

## Configuration

Set these environment variables if needed:

```bash
export MONGODB_URI="mongodb://localhost:27017"
export REDIS_URL="redis://localhost:6379/0"
export API_KEY="your-custom-key"  # Default: test-api-key
export HTTP_PORT="8080"
```

## API Endpoints

All requests require the `Authorization` header with your API key(deafualt: "test-api-key").

### 1. Get Gross Gaming Revenue (GGR)

Calculate casino profit across different currencies.

```
curl -H "Authorization:test-api-key" "http://localhost:8080/gross_gaming_rev?from=2023-01-01T00:00:00Z&to=2023-12-31T23:59:59Z
```

**Example Response:**
```json
{
  "timeframe": {
    "from": "2023-01-01T00:00:00Z",
    "to": "2023-12-31T23:59:59Z"
  },
  "data": [
    {
      "currency": "BTC",
      "ggr": "15.23",
      "ggrUSD": "761500.00"
    },
    {
      "currency": "ETH",
      "ggr": "105.75",
      "ggrUSD": "211500.00"
    },
    {
      "currency": "USDT",
      "ggr": "26541.50",
      "ggrUSD": "26541.50"
    }
  ]
}
```

### 2. Get Daily Wager Volume

See how much players bet each day by currency.

```
curl -H "Authorization:test-api-key" "http://localhost:8080/daily_wager_volume?from=2023-01-01T00:00:00Z&to=2023-01-07T23:59:59Z
```

**Example Response:**
```json
{
  "timeframe": {
    "from": "2023-01-01T00:00:00Z",
    "to": "2023-01-07T23:59:59Z"
  },
  "data": [
    {
      "date": "2023-01-01",
      "currency": "BTC",
      "wagerAmount": "12.45",
      "wagerUSDAmount": "622500.00"
    },
    {
      "date": "2023-01-01",
      "currency": "ETH",
      "wagerAmount": "150.75",
      "wagerUSDAmount": "301500.00"
    },
    {
      "date": "2023-01-02",
      "currency": "BTC",
      "wagerAmount": "9.33",
      "wagerUSDAmount": "466500.00"
    }
  ]
}
```

### 3. Get User Wager Percentile

Find where a player ranks compared to others (e.g., top 2%).

```
curl -H "Authorization:test-api-key" "http://localhost:8080/user/01HRMD5HGTZB3TW3PGYXRD07CQT/wager_percentile?from=2023-01-01T00:00:00Z&to=2023-12-31T23:59:59Z
```

**Example Response:**
```json
{
  "userID": "01HRMD5HGTZB3TW3PGYXRD07CQT",
  "percentile": 97.5,
  "timeframe": {
    "from": "2023-01-01T00:00:00Z",
    "to": "2023-12-31T23:59:59Z"
  }
}
```

## Docker Setup

To run everything in Docker:

```bash
# Create docker-compose.yml
cat > docker-compose.yml << EOF
version: '3'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=mongodb://mongodb:27017
      - REDIS_URL=redis://redis:6379/0
    depends_on:
      - mongodb
      - redis
  
  mongodb:
    image: mongo:latest
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
  
  redis:
    image: redis:latest
    ports:
      - "6379:6379"

volumes:
  mongodb_data:
EOF

# Create Dockerfile
cat > Dockerfile << EOF
FROM golang:1.19-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o server ./cmd/api

EXPOSE 8080
CMD ["./server"]
EOF

# Run with Docker Compose
docker-compose up -d

# Run seed utility
go run cmd/seed/main.go
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover
```