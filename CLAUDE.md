# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Ironman is a Go-based construction safety inspection application that uses AI to analyze photos for safety violations. The application follows a standard HTTP server pattern with HTML templates and includes both public-facing and authenticated app sections.

## Architecture

### Core Structure
- **Entry Point**: `client/main.go` - Standard Go HTTP server with graceful shutdown
- **Configuration**: Environment-based config with flag overrides (client/config.go)
- **Database**: PostgreSQL using pgx/v5 driver
- **Web Framework**: Echo v4 for routing and middleware
- **Templates**: Standard Go html/template with custom helper functions
- **Static Assets**: Served from `client/public/static/`

### Key Components
- **DTOs**: Data transfer objects in `dto/dto.go` for construction projects, violations, users
- **Templates**: HTML templates in `client/public/views/` organized by section (auth/, app/, layout/)
- **Template Helpers**: Custom template functions in `client/public/views/template_helpers.go`
- **Routes**: HTTP routes defined in `client/routes.go` with mock data functions
- **Middleware**: Request logging and other middleware in `client/middleware.go`

### Application Domains
- **Construction Projects**: Project management with compliance tracking
- **Safety Violations**: AI-detected safety issues with OSHA regulation references
- **User Management**: Role-based access (inspector, admin) with company associations
- **Reporting**: Timeline events, compliance scores, photo management

## Development Commands

### Running the Application
```bash
# Development mode
go run client/*.go

# With custom configuration
go run client/*.go --app_host=0.0.0.0 --app_port=3000 --log_level=debug

# Environment variables can also be used:
# DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
# APP_HOST, APP_PORT, LOG_LEVEL, ENVIRONMENT, ANTHROPIC_API_KEY
```

### Database
- Requires PostgreSQL connection
- Default connection: `postgres://postgres:@localhost:5432/postgres`
- Database connection established at startup with ping verification

### Build and Test
```bash
# Build
go build -o ironman client/*.go

# Standard Go testing
go test ./...

# Dependencies
go mod tidy
```

## Code Patterns

### Configuration Precedence
Configuration follows: default → environment variables → command line flags (highest priority)

### Template Structure
- Layout templates in `layout/`
- Feature templates in `auth/`, `app/`, `public/`
- Templates use custom helper functions for status badges, pagination, etc.

### Data Flow
Mock data functions in `client/routes.go` simulate database operations. These should be replaced with actual database queries when implementing persistence.

### Logging
Structured logging using slog with JSON output in production, text output in development.