# Car Listing Service

A production-ready Go REST API service for scraping and managing Facebook Marketplace car listings, built with Gin framework and ChromeDP.

## Project Summary

I found this challenge incredibly interesting and learned extensively throughout the development process. Here are the key technical challenges I tackled:

### Core Technical Challenges Solved

**1. Facebook Authentication Approach**
- Initially attempted direct login via web APIs but discovered Facebook's JavaScript-heavy authentication prevents standard API-based login
- Pivoted to using **mbasic.facebook.com** approach with cookie persistence
- Eventually upgraded to **ChromeDP headless browser** automation for reliable authentication
- Implemented cookie caching system to avoid repeated logins

**2. Infinite Scroll Marketplace Challenge**
- Facebook Marketplace uses infinite scroll, making complete data extraction extremely challenging
- Developed sophisticated **scroll state detection** algorithm with multiple stop signals:
  - DOM unchanged detection
  - Consecutive no-new-items tracking
  - Scroll position monitoring
- Implemented **adaptive delay mechanism** that adjusts based on:
  - New items discovery rate
  - Duplicate detection ratio
  - Network performance

**3. Batched Data Processing**
- Implemented **channel-based streaming** to prevent in-memory data loss if scraper crashes
- Cars are processed in batches and immediately persisted to PostgreSQL
- Prevents memory overflow on large datasets

### Production-Ready Implementation

**Architecture & Code Quality:**
- Full **MVC architecture** (Models, Controllers, Services, Repository pattern)
- Environment-based configuration via `.env` with sensible defaults
- Docker Compose setup for PostgreSQL
- Graceful shutdown handling
- Structured logging with progress tracking
- Clean code: removed all unnecessary comments and logs
- Database migrations for schema management

**Key Features:**
- Cookie persistence for session reuse
- Configurable scraper parameters (max scrolls, delays, timeouts)
- Duplicate detection using URL-based deduplication
- Real-time progress logging (items/minute, duplicate ratios)
- RESTful API endpoints for CRUD operations
- Health check endpoint for monitoring

## Project Structure

```
.
├── config/          # Environment & scraper configuration
├── controllers/     # HTTP request handlers
├── database/        # Database connection
├── middleware/      # CORS, Logger middleware
├── migrations/      # SQL migrations
├── models/          # Data models
├── repository/      # Data access layer
├── routes/          # API routes
├── services/        # Business logic & Facebook scraper
└── main.go          # Entry point with graceful shutdown
```

## Setup

1. Copy `.env.example` to `.env` and configure your environment variables:
```bash
cp .env.example .env
```

2. Start PostgreSQL with Docker Compose:
```bash
docker-compose up -d
```

3. Install dependencies:
```bash
go mod download
```

4. Run database migrations:
```bash
# Migrations are in migrations/ directory
# Apply them to your database
```

5. Run the application:
```bash
go run main.go
```

## Environment Variables

### Server Configuration
- `SERVER_PORT`: Server port (default: 3001)
- `ENVIRONMENT`: Environment mode (development/production)

### Database Configuration
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name (default: car_listing)

### Facebook Authentication
- `FACEBOOK_EMAIL`: Your Facebook email for scraper authentication
- `FACEBOOK_PASSWORD`: Your Facebook password for scraper authentication

### Scraper Configuration
- `SCRAPER_MAX_SCROLLS`: Maximum scroll iterations (default: 2000)
- `SCRAPER_MAX_DURATION`: Maximum scraping duration (default: 60m)
- `SCRAPER_INITIAL_DELAY`: Initial delay between scrolls (default: 2s)
- `SCRAPER_MIN_DELAY`: Minimum delay between scrolls (default: 1500ms)
- `SCRAPER_MAX_DELAY`: Maximum delay between scrolls (default: 5s)
- `SCRAPER_MAX_CONSECUTIVE_NO_NEW`: Max scrolls with no new items before stopping (default: 10)
- `SCRAPER_MAX_CONSECUTIVE_UNCHANGED`: Max scrolls with unchanged DOM before stopping (default: 10)
- `SCRAPER_EXTRACTION_INTERVAL`: Log progress every N scrolls (default: 5)

## API Endpoints

### Health Check
- `GET /health` - Service health check

### Car Listings
- `GET /api/v1/cars` - Get all car listings
- `GET /api/v1/cars/:id` - Get car by ID
- `POST /api/v1/cars` - Create new car listing
- `PUT /api/v1/cars/:id` - Update car listing
- `DELETE /api/v1/cars/:id` - Delete car listing
- `POST /api/v1/cars/scrape` - Trigger Facebook Marketplace scraping

## Features

- Production-ready MVC architecture
- Facebook authentication with cookie persistence
- Intelligent infinite scroll handling
- Real-time scraping progress monitoring
- Batched data processing with crash recovery
- Docker Compose for database
- Environment-based configuration
- Custom middleware (Logger, CORS)
- Graceful shutdown
- Structured error handling
- Adaptive scraping delays
- Duplicate detection and deduplication
