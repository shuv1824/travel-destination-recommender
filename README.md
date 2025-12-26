# Travel Destination Recommender API

An intelligent travel recommendation API that helps users make informed travel decisions based on real-time weather conditions and air quality data across Bangladesh.

## Problem Statement

Planning a trip within Bangladesh can be challenging when you want to escape extreme heat and poor air quality. This API solves that problem by:

- Analyzing real-time temperature and air quality (PM2.5) data for all districts
- Comparing your current location conditions with potential destinations
- Providing data-driven recommendations for cooler and cleaner travel destinations
- Helping you discover the top destinations with the best weather conditions

## Features

- **Smart Travel Recommendations** - Get personalized travel advice based on temperature and air quality comparisons
- **Top 10 Destinations** - Discover the coolest and cleanest districts in Bangladesh
- **7-Day Weather Forecasts** - Real-time data from Open-Meteo weather APIs
- **Air Quality Monitoring** - Track PM2.5 pollution levels for health-conscious travel
- **High Performance** - Concurrent API calls with intelligent caching (5-minute TTL)
- **Background Refresh** - Automatic cache updates every 2.5 minutes
- **Production Ready** - Graceful shutdown, health checks, CORS support, and comprehensive error handling

## Quick Start

### Prerequisites

- Go 1.23.6 or later
- Docker (optional)

### Installation & Running

```bash
# Clone the repository
git clone https://github.com/shuv1824/recommender.git
cd travel-dest-rec

# Build the application
go build -o recommender .

# Run the server
./recommender
```

The server will start on `http://localhost:8080`. Initial startup takes ~60 seconds to warm the cache.

### Using Docker

```bash
# Build and run with Docker Compose (includes live reload)
docker-compose up

# Or build manually
docker build -t travel-recommender .
docker run -p 8080:8080 travel-recommender
```

### Quick Example

```bash
# Get top 10 coolest and cleanest destinations
curl http://localhost:8080/api/v1/destinations/top

# Get travel recommendation
curl -X POST http://localhost:8080/api/v1/travel/recommendation \
  -H "Content-Type: application/json" \
  -d '{
    "current_location": {
      "lat": 23.8103,
      "long": 90.4125,
      "name": "Dhaka"
    },
    "destination_district": "Cox'\''s Bazar",
    "travel_date": "2025-12-27"
  }'
```

## API Documentation

### Base URL

```
http://localhost:8080
```

### Endpoints

#### 1. Health Check

Check if the service is running.

```http
GET /health
```

**Response (200 OK):**

```json
{
  "data": {
    "status": "healthy"
  }
}
```

---

#### 2. Get Top Destinations

Returns the top 10 coolest and cleanest districts based on 7-day forecast averages.

```http
GET /api/v1/destinations/top
```

**Response Headers:**

- `X-Response-Time`: Request execution time in milliseconds

**Response (200 OK):**

```json
{
  "data": {
    "generated_at": "2025-12-26T12:34:56Z",
    "description": "Top 10 coolest and cleanest districts in Bangladesh based on 7-day forecast (2PM temperature and PM2.5 levels)",
    "destinations": [
      {
        "id": "1",
        "name": "Sylhet",
        "avg_temp_2pm_celsius": 24.5,
        "avg_pm25": 28.3,
        "rank": 1
      },
      {
        "id": "2",
        "name": "Cox's Bazar",
        "avg_temp_2pm_celsius": 25.2,
        "avg_pm25": 30.1,
        "rank": 2
      }
      // ... 8 more destinations
    ]
  }
}
```

**Ranking Logic:**

1. Sorted by average 2PM temperature (ascending)
2. Ties broken by average PM2.5 levels (ascending)
3. Returns top 10 districts

**Error Responses:**

- `504 Gateway Timeout` - Request exceeded 490ms timeout
- `500 Internal Server Error` - Weather service unavailable

**Example:**

```bash
curl -X GET http://localhost:8080/api/v1/destinations/top
```

---

#### 3. Get Travel Recommendation

Get a personalized travel recommendation comparing your current location with a destination.

```http
POST /api/v1/travel/recommendation
Content-Type: application/json
```

**Request Body:**

```json
{
  "current_location": {
    "lat": 23.8103,
    "long": 90.4125,
    "name": "Dhaka"
  },
  "destination_district": "Cox's Bazar",
  "travel_date": "2025-12-27"
}
```

**Request Parameters:**

| Field                   | Type    | Required | Description                                                 |
| ----------------------- | ------- | -------- | ----------------------------------------------------------- |
| `current_location.lat`  | float64 | Yes      | Latitude of current location                                |
| `current_location.long` | float64 | Yes      | Longitude of current location                               |
| `current_location.name` | string  | No       | Name of current location                                    |
| `destination_district`  | string  | Yes      | Name of destination district (must exist in districts.json) |
| `travel_date`           | string  | Yes      | Travel date in YYYY-MM-DD format (within next 7 days)       |

**Response (200 OK):**

```json
{
  "data": {
    "recommendation": "Recommended",
    "reason": "Cox's Bazar is significantly cooler (7.5°C less) and has significantly better air quality. Enjoy your trip!",
    "travel_date": "2025-12-27",
    "current_location": {
      "name": "Dhaka",
      "temp_2pm_celsius": 35.0,
      "pm25": 75.0
    },
    "destination": {
      "name": "Cox's Bazar",
      "temp_2pm_celsius": 27.5,
      "pm25": 25.0
    },
    "temp_difference_celsius": 7.5,
    "pm25_difference": 50.0
  }
}
```

**Recommendation Values:**

- `"Recommended"` - Destination is both cooler AND cleaner than current location
- `"Not Recommended"` - Destination is either hotter or has worse air quality

**Reason Messages:**

The API generates human-readable reasons based on:

**Temperature Difference:**

- `< 1°C` - "about the same temperature"
- `1-3°C` - "slightly cooler/hotter"
- `> 3°C` - "significantly cooler/hotter"

**Air Quality Difference (PM2.5):**

- `< 5` - "similar air quality"
- `5-15` - "better/worse air quality"
- `> 15` - "significantly better/worse air quality"

**Error Responses:**

**400 Bad Request:**

```json
{
  "error": {
    "code": 400,
    "message": "current_location lat and long are required"
  }
}
```

**Common Errors:**

- Missing `lat` or `long` in current_location
- Missing `destination_district`
- Missing `travel_date`
- Invalid date format (must be YYYY-MM-DD)
- Travel date in the past
- Travel date beyond 7-day forecast window
- Destination district not found in database

**Examples:**

```bash
# Recommended destination (cooler and cleaner)
curl -X POST http://localhost:8080/api/v1/travel/recommendation \
  -H "Content-Type: application/json" \
  -d '{
    "current_location": {
      "lat": 23.8103,
      "long": 90.4125,
      "name": "Dhaka"
    },
    "destination_district": "Cox'\''s Bazar",
    "travel_date": "2025-12-27"
  }'

# Check if Sylhet is a good destination
curl -X POST http://localhost:8080/api/v1/travel/recommendation \
  -H "Content-Type: application/json" \
  -d '{
    "current_location": {
      "lat": 23.8103,
      "long": 90.4125
    },
    "destination_district": "Sylhet",
    "travel_date": "2025-12-28"
  }'
```

## Project Structure

```
travel-dest-rec/
├── cmd/
│   └── root.go                      # Application initialization & server setup
├── internal/
│   ├── handler/
│   │   └── handler.go               # HTTP request handlers
│   ├── services/
│   │   ├── weather/
│   │   │   ├── service.go           # Weather API integration
│   │   │   ├── cached_service.go    # Caching layer with background refresh
│   │   │   └── service_test.go      # Weather service tests
│   │   └── travel/
│   │       ├── service.go           # Travel recommendation logic
│   │       └── service_test.go      # Travel service tests
│   ├── types/
│   │   └── types.go                 # Type definitions (DTOs, models)
│   ├── utils/
│   │   └── geodata/
│   │       └── geodata.go           # District data loading utility
│   └── response/
│       └── response.go              # HTTP response helpers
├── data/
│   └── districts.json               # Bangladesh district coordinates (64 districts)
├── main.go                          # Application entry point
├── go.mod                           # Go module definition
├── Dockerfile                       # Multi-stage Docker build (Alpine-based)
├── docker-compose.yaml              # Development environment with live reload
└── .air.toml                        # Live reload configuration
```

### Component Responsibilities

- **cmd/**: Server initialization, routing, middleware setup, graceful shutdown
- **internal/handler/**: Maps HTTP routes to service layer, request validation, response formatting
- **internal/services/weather/**: Fetches and caches weather + air quality data from Open-Meteo APIs
- **internal/services/travel/**: Business logic for travel recommendations and reason generation
- **internal/types/**: Shared data structures across layers
- **internal/utils/geodata/**: Loads and provides access to Bangladesh district data
- **internal/response/**: Standardized JSON response formatting
- **data/**: Static datasets (district coordinates and metadata)

## Development

### Prerequisites

- Go 1.23.6 or later
- Docker & Docker Compose (for containerized development)
- [Air](https://github.com/cosmtrek/air) (optional, for live reload)

### Local Development

```bash
# Install dependencies
go mod download

# Run with live reload (requires Air)
air

# Or run directly
go run main.go

# Build for production
go build -ldflags="-w -s" -o recommender .
```

### Development with Docker

```bash
# Start with live reload
docker-compose up

# Rebuild after dependency changes
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/services/weather/
go test ./internal/services/travel/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Coverage:**

- **Weather Service Tests** (`internal/services/weather/service_test.go`)
  - District ranking logic
  - Cache hit/miss scenarios
  - Cache expiry and refresh
  - Data integrity (copy vs. reference)

- **Travel Service Tests** (`internal/services/travel/service_test.go`)
  - Successful recommendations (recommended/not recommended)
  - Date validation (format, range, past dates)
  - District validation
  - Reason generation for various scenarios
  - Mock HTTP client for external API calls

## Configuration

### Server Configuration

| Setting                   | Value | Description                           |
| ------------------------- | ----- | ------------------------------------- |
| Port                      | 8080  | HTTP server port                      |
| Graceful Shutdown Timeout | 30s   | Time to complete in-flight requests   |
| Top Destinations Timeout  | 490ms | Request timeout for /destinations/top |

### Cache Configuration

| Setting                     | Value       | Description                                     |
| --------------------------- | ----------- | ----------------------------------------------- |
| Cache TTL                   | 5 minutes   | How long cached data remains valid              |
| Background Refresh Interval | 2.5 minutes | How often cache is refreshed in background      |
| Warm Cache on Startup       | Yes         | Pre-populate cache on server start (takes ~60s) |
| Warm Cache Timeout          | 60s         | Max time for initial cache warming              |

### Weather Service Configuration

| Setting                  | Value       | Description                                |
| ------------------------ | ----------- | ------------------------------------------ |
| Max Concurrent API Calls | 5           | Semaphore limit for parallel requests      |
| HTTP Client Timeout      | 10s         | Timeout for each external API call         |
| Data Point               | 2PM (14:00) | Time of day for temperature/PM2.5 readings |

### Middleware Stack

1. **Recovery Handler** - Catches panics and returns 500 errors
2. **CORS Handler** - Allows cross-origin requests (all origins, GET/POST methods)
3. **Logging Handler** - Logs requests to stdout

### Data Files

The application requires `data/districts.json` containing all Bangladesh districts:

```json
{
  "districts": [
    {
      "id": "1",
      "division_id": "3",
      "name": "Dhaka",
      "bn_name": "ঢাকা",
      "lat": "23.7115253",
      "long": "90.4111451"
    }
    // ... 63 more districts
  ]
}
```

## External APIs

### Open-Meteo Weather Forecast API

- **URL:** `https://api.open-meteo.com/v1/forecast`
- **Purpose:** Hourly temperature forecasts for 7 days
- **Parameters:** latitude, longitude, hourly=temperature_2m, timezone=auto
- **Authentication:** None required

### Open-Meteo Air Quality API

- **URL:** `https://air-quality-api.open-meteo.com/v1/air-quality`
- **Purpose:** Hourly PM2.5 (particulate matter) levels for 7 days
- **Parameters:** latitude, longitude, hourly=pm2_5, timezone=auto
- **Authentication:** None required

Both APIs are:

- Free to use
- No API key required
- High availability
- Updated hourly
- Cover global locations

## Architecture Highlights

### Performance Optimizations

1. **Concurrent API Calls** - Fetches temperature and PM2.5 data in parallel using goroutines
2. **Intelligent Caching** - 5-minute cache with background refresh to minimize API calls
3. **Semaphore Pattern** - Limits concurrent requests to 5 to avoid overwhelming external APIs
4. **HTTP Connection Pooling** - Reuses connections with `MaxIdleConns=100`
5. **Warm Cache on Startup** - Pre-fetches data before serving requests

### Design Patterns

- **Service Layer Architecture** - Separation of HTTP handlers and business logic
- **Repository Pattern** - `geodata` utility abstracts district data access
- **Decorator Pattern** - `CachedWeatherService` wraps `WeatherService` to add caching
- **Concurrent Pipeline** - Goroutines + channels for parallel data fetching

### Error Handling

- Comprehensive validation of request inputs
- Graceful handling of external API failures
- Structured error responses with HTTP status codes
- Panic recovery middleware

## API Response Format

All successful responses follow this format:

```json
{
  "data": {
    // ... response data
  }
}
```

Error responses use:

```json
{
  "error": {
    "code": 400,
    "message": "descriptive error message"
  }
}
```

## Limitations

- **Forecast Range:** Only supports 7-day forecasts (Open-Meteo API limitation)
- **Geographic Scope:** Currently limited to Bangladesh districts
- **Data Point:** Uses 2PM temperature (may not represent full day conditions)
- **Cache Staleness:** Up to 5 minutes of stale data possible
- **Rate Limits:** Dependent on Open-Meteo free tier limits

---

Built with Go and powered by Open-Meteo APIs
