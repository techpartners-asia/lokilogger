# Lokilogger

A Go package that provides structured logging with both local logging (using Zap) and remote logging to Grafana Loki.

## Features

- Structured logging using Zap
- Remote logging to Grafana Loki
- Multiple log levels (Info, Error, Debug, Warn)
- Configurable service and environment tags
- JSON encoding for better parsing
- Error handling and context
- Middleware examples for popular web frameworks (Gin, Fiber)
- Request/Response logging with detailed metrics
- Automatic error tracking and monitoring

## Installation

```bash
go get github.com/techpartners-asia/lokilogger
```

## Quick Start

```go
package main

import (
    "github.com/techpartners-asia/lokilogger"
    "go.uber.org/zap"
)

func main() {
    // Create logger configuration
    config := lokilogger.Config{
        BaseURL:     "http://your-loki-server:3100",
        Environment: "development",
        Service:     "your-service",
    }

    // Create new logger instance
    logger, err := lokilogger.New(config)
    if err != nil {
        panic(err)
    }

    // Log with structured fields
    fields := []zap.Field{
        zap.String("user_id", "123"),
        zap.String("action", "login"),
    }

    // Info level logging
    logger.Info("User logged in successfully", fields...)

    // Error level logging
    logger.Error("Failed to process request", err, fields...)

    // Debug level logging
    logger.Debug("Processing request details", fields...)

    // Warn level logging
    logger.Warn("High memory usage detected", fields...)
}
```

## RequestLogger Example

The RequestLogger middleware provides automatic request logging with detailed metrics. Here's how to implement it:

### Gin RequestLogger

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/techpartners-asia/lokilogger"
    "go.uber.org/zap"
)

// RequestLogger middleware for Gin
func RequestLogger(logger *lokilogger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery

        // Process request
        c.Next()

        // Log after request is processed
        latency := time.Since(start)
        statusCode := c.Writer.Status()

        fields := []zap.Field{
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.Int("status", statusCode),
            zap.Duration("latency", latency),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        }

        // Log based on status code
        switch {
        case statusCode >= 500:
            logger.Error("Server error", nil, fields...)
        case statusCode >= 400:
            logger.Warn("Client error", fields...)
        default:
            logger.Info("Request processed", fields...)
        }
    }
}
```

### Fiber RequestLogger

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/techpartners-asia/lokilogger"
    "go.uber.org/zap"
)

// RequestLogger middleware for Fiber
func RequestLogger(logger *lokilogger.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        path := c.Path()
        query := c.Queries()

        // Process request
        err := c.Next()
        if err != nil {
            return err
        }

        // Log after request is processed
        latency := time.Since(start)
        statusCode := c.Response().StatusCode()

        fields := []zap.Field{
            zap.String("method", c.Method()),
            zap.String("path", path),
            zap.Any("query", query),
            zap.Int("status", statusCode),
            zap.Duration("latency", latency),
            zap.String("client_ip", c.IP()),
            zap.String("user_agent", c.Get("User-Agent")),
        }

        // Log based on status code
        switch {
        case statusCode >= 500:
            logger.Error("Server error", nil, fields...)
        case statusCode >= 400:
            logger.Warn("Client error", fields...)
        default:
            logger.Info("Request processed", fields...)
        }

        return nil
    }
}
```

### Using RequestLogger

```go
// Initialize logger
config := lokilogger.Config{
    BaseURL:     "http://your-loki-server:3100",
    Environment: "development",
    Service:     "your-service",
}
logger, _ := lokilogger.New(config)

// For Gin
router := gin.New()
router.Use(RequestLogger(logger))

// For Fiber
app := fiber.New()
app.Use(RequestLogger(logger))
```

The RequestLogger middleware automatically captures:
- Request method and path
- Query parameters
- Status code and latency
- Client information
- Error details
- Performance metrics

## Web Framework Integration

### Gin Example

The package includes a middleware for Gin that automatically logs HTTP requests. Here's a complete example:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/techpartners-asia/lokilogger"
)

func main() {
    // Initialize logger
    config := lokilogger.Config{
        BaseURL:     "http://your-loki-server:3100",
        Environment: "development",
        Service:     "gin-example",
    }
    logger, _ := lokilogger.New(config)

    // Create Gin router with logging middleware
    router := gin.New()
    router.Use(RequestLogger(logger))

    // Example routes
    router.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello from Gin!"})
    })

    router.GET("/user/:id", func(c *gin.Context) {
        userID := c.Param("id")
        logger.Info("User profile accessed",
            zap.String("user_id", userID),
            zap.String("action", "profile_view"),
        )
        c.JSON(200, gin.H{"user_id": userID})
    })

    router.Run(":8080")
}
```

### Fiber Example

The package also includes a middleware for Fiber. Here's a complete example:

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/techpartners-asia/lokilogger"
)

func main() {
    // Initialize logger
    config := lokilogger.Config{
        BaseURL:     "http://your-loki-server:3100",
        Environment: "development",
        Service:     "fiber-example",
    }
    logger, _ := lokilogger.New(config)

    // Create Fiber app with logging middleware
    app := fiber.New()
    app.Use(RequestLogger(logger))

    // Example routes
    app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "Hello from Fiber!",
        })
    })

    app.Get("/user/:id", func(c *fiber.Ctx) error {
        userID := c.Params("id")
        logger.Info("User profile accessed",
            zap.String("user_id", userID),
            zap.String("action", "profile_view"),
        )
        return c.JSON(fiber.Map{
            "user_id": userID,
        })
    })

    app.Listen(":8080")
}
```

## Request Logging Features

The middleware automatically logs the following information for each request:

### Basic Information
- Request method (GET, POST, etc.)
- Request path
- Query parameters
- Status code
- Response latency
- Client IP address
- User agent

### Log Levels Based on Status Code
- 5xx: Error level (server errors)
- 4xx: Warning level (client errors)
- 2xx/3xx: Info level (successful requests)

### Additional Features
- Structured logging with Zap
- JSON encoding for better parsing
- Automatic error tracking
- Performance monitoring
- Request tracing

## Configuration

The `Config` struct allows you to configure the logger:

```go
type Config struct {
    BaseURL     string // Loki server URL
    Environment string // Environment (e.g., "development", "production")
    Service     string // Service name
}
```

## Log Levels

The package supports the following log levels:

- `Info`: For general information
- `Error`: For error conditions (includes error details)
- `Debug`: For detailed debugging information
- `Warn`: For warning conditions

## Structured Fields

You can add structured fields to your logs using Zap's field constructors:

```go
fields := []zap.Field{
    zap.String("key", "value"),
    zap.Int("count", 42),
    zap.Time("timestamp", time.Now()),
    zap.Error(err),
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License 