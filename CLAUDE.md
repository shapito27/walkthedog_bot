# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

WalkTheDog Bot is a Telegram bot (@walkthedog_bot) that helps people sign up for trips to animal shelters in the Moscow region. The bot collects user registration data and stores it in Google Sheets for the organizers.

## Development Commands

### Running the Bot
- **Development**: `make run_app` or `go run main.go`
- **Tests**: `make dev/test` or `go test`

### Docker Operations
- **Build**: `make docker/build`
- **Create container**: `make docker/create_container`
- **Start/Stop/Restart**: `make docker/container/start|stop|restart`
- **Logs**: `make docker/logs`
- **Clean up**: `make docker/clean`

## Architecture

### Core Structure
- **main.go**: Main bot logic with state management and command handling
- **internal/models/**: Data models for shelters, trips, and configuration
- **internal/google/sheet/**: Google Sheets integration for data storage
- **internal/dates/**: Date utilities (Russian weekday names)
- **configs/**: Configuration files for app settings and shelter information

### Key Components

**Bot State Management**:
- Each chat maintains state in `statePool` map
- State includes last message and current trip registration
- Commands flow through a state machine pattern

**Shelter System**:
- Shelters configured in `configs/shelters.yml`
- Each shelter has schedule (regularly/everyday/none types)
- Schedule supports week/day patterns with exceptions

**Google Sheets Integration**:
- Requires `credentials.json` and `token.json` for OAuth
- Data saved to separate sheets per shelter + System sheet
- Includes caching mechanism for offline resilience

### Configuration Files

**configs/app.yml**:
- Telegram bot tokens for different environments
- Admin settings and Google Sheets configuration
- Copy from `configs/app.yml.example` and configure

**configs/shelters.yml**:
- Complete shelter database with schedules, links, and metadata
- Schedule types: "regularly" (specific weeks/days), "everyday", "none"
- Each shelter has donation links and capacity limits

### Admin Commands
- `/reread_shelters`: Reload shelter configuration
- `/reread_app_config`: Reload app configuration  
- `/update_google_auth`: Update Google Sheets authentication
- `/clear_cache`: Send cached trips to sheets and clear cache

### User Flow
1. Start with `/start` - shows main menu
2. Choose `/go_shelter` - select by shelter or date
3. Registration flow: shelter → date → first time → purposes → transport → source
4. Data saved to Google Sheets with offline caching fallback

## Development Notes

- Go 1.18+ required
- Uses Telegram Bot API v5
- Implements in-memory caching with file persistence
- State machine architecture for conversation flow
- Shelter scheduling supports complex recurring patterns
- Google OAuth2 integration for sheets access