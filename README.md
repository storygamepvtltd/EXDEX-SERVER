# Project Structure

This document outlines the folder structure of the application.

```
/Users/mac/Documents/storyGame/exdex new part/
├───.gitignore
├───app.yml
├───go.mod
├───go.sum
├───main.go
├───Makefile.cp
├───run.sh
├───cmd/
├───config/
├───docker/
├───indexing/
├───internal/
│   ├───app/
│   │   └───app.go
│   ├───router/
│   │   └───main_router.go
│   └───src/
│       ├───handler/
│       ├───model/
│       │   └───timestamps.go
│       ├───repository/
│       │   ├───interfaces.go
│       │   └───repo.go
│       ├───seed/
│       └───services/
├───server/
│   ├───config/
│   ├───constant/
│   ├───cron/
│   ├───databases/
│   │   └───mongo.go
│   ├───info/
│   │   └───info.go
│   ├───jwt/
│   │   └───jwt.go
│   ├───middleware/
│   ├───queue/
│   ├───responses/
│   │   └───respons,go
│   ├───security/
│   │   └───security.go
│   ├───static/
│   ├───utils/
│   └───validator/
│       └───validator.go
├───test/
└───uploads/
```

## Root Directory

*   `.gitignore`: Specifies files and folders to be ignored by Git.
*   `app.yml`: Main configuration file for the application.
*   `go.mod` & `go.sum`: Manage the project's dependencies.
*   `main.go`: The entry point of the application.
*   `Makefile.cp`: Makefile for copy operations.
*   `run.sh`: Script for running the application.

## Directories

*   `cmd/`: Contains the main application entry point.
*   `config/`: Handles application configuration loading.
*   `docker/`: Stores Docker-related files (e.g., Dockerfile).
*   `indexing/`: Contains logic related to data indexing (e.g., Elasticsearch).
*   `internal/`: Houses the core application logic.
    *   `app/`: Core application setup and initialization.
    *   `router/`: Defines the application's API routes.
    *   `src/`: Contains the main source code.
        *   `handler/`: HTTP request handlers.
        *   `model/`: Data models and structures.
        *   `repository/`: Data access layer for interacting with the database.
        *   `seed/`: Scripts for seeding the database with initial data.
        *   `services/`: Business logic and services.
*   `server/`: Contains server-related components.
    *   `config/`: Server-specific configuration.
    *   `constant/`: Application-wide constants.
    *   `cron/`: Cron job definitions and scheduling.
    *   `databases/`: Database connection and management.
    *   `info/`: Server information and health checks.
    *   `jwt/`: JWT generation, validation, and authentication.
    *   `middleware/`: Custom HTTP middleware.
    *   `queue/`: Message queue implementation.
    *   `responses/`: Standardized API response structures.
    *   `security/`: Security-related utilities.
    *   `static/`: Static assets (e.g., images, CSS, JavaScript).
    *   `utils/`: Utility functions.
    *   `validator/`: Request data validation.
*   `test/`: Contains test files for the application.
*   `uploads/`: Directory for storing uploaded files.
