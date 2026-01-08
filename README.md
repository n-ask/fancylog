<p align="center">
  <picture>
    <img alt="fancylog: An extremely fast golang logger" src="./assets/fancylog.png" width="300">
  </picture>
</p>
<p>
fancylog provides stylized, non-blocking log output. It is designed to be fast, visually appealing, and highly configurable. It is used across many Golang projects here at NASK.

### Features

- **Extremely Fast**: Optimized for performance with minimal overhead.
- **Non-blocking**: Designed to handle high-throughput logging without slowing down your application.
- **Visual Clarity**: Uses ANSI colors to distinguish between different log levels and components.
- **Structured Logging**: Supports logging maps and structured data for better machine readability.
- **Specialized HTTP Logging**: Includes a dedicated logger for HTTP requests with status code coloring.
- **Middleware Support**: Easy integration with Gin, gRPC, pgx, and standard `net/http`.
- **Customizable**: Configure names, timestamps, colors, and more.

<p align="center">
  <picture>
    <img alt="fancylog:example" src="./assets/fancylog.gif" width="700">
  </picture>
</p>

### Installation

```bash
go get github.com/n-ask/fancylog
```

### Quick Start

#### Basic Logging

```go
package main

import (
    "os"
    "github.com/n-ask/fancylog"
)

func main() {
    // Create a new logger with os.Stdout as output
    log := fancylog.New(os.Stdout)

    log.Info("Starting application...")
    log.Warn("This is a warning message.")
    log.Error("Something went wrong!")
    
    // Structured logging with maps
    log.InfoMap(map[string]any{
        "event": "user_login",
        "user_id": 123,
        "status": "success",
    })
}
```

#### HTTP Logging

```go
package main

import (
    "os"
    "github.com/n-ask/fancylog"
)

func main() {
    // Create a specialized HTTP logger
    httpLog := fancylog.NewHttpLogger(os.Stdout)

    // Log an HTTP GET request with a 200 OK status
    httpLog.GetMethod(map[string]any{
        "uri": "/api/v1/users",
        "ip": "127.0.0.1",
    }, 200)

    // Log an error status
    httpLog.PostMethod(map[string]any{
        "uri": "/api/v1/login",
    }, 401)
}
```

### Integration with Frameworks

#### Gin Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/n-ask/fancylog"
    "github.com/n-ask/fancylog/handlers"
)

r := gin.New()
log := fancylog.NewHttpLogger(os.Stdout)
r.Use(handlers.GinLogger(log))
```

#### gRPC Interceptors

```go
import (
    "google.golang.org/grpc"
    "github.com/n-ask/fancylog"
    "github.com/n-ask/fancylog/handlers"
)

log := fancylog.New(os.Stdout)
s := grpc.NewServer(
    grpc.UnaryInterceptor(handlers.UnaryServerInterceptor(log, fancylog.Info, myLogFunc, nil)),
)
```

### Configuration

`FancyLogger` provides several methods to customize its behavior:

- `WithColor()` / `WithoutColor()`: Enable or disable ANSI colors.
- `WithDebug()` / `WithoutDebug()`: Show or hide debug and trace level logs.
- `WithTimestamp()` / `WithoutTimestamp()`: Enable or disable timestamps.
- `Quiet()` / `NoQuiet()`: Silence all output or resume logging.
- `SetTimestampColor(color)`: Customize the color of the timestamp.
- `NewWithName(name, output)`: Create a logger with a specific prefix name.
