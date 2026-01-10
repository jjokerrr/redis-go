# Redis-Go Project Context

## Project Overview

Redis-Go is a simple implementation of a Redis-compatible server written in Go, primarily designed for learning purposes. The project aims to replicate core Redis functionality while providing an educational resource for understanding how Redis-like systems work.

The project is structured as a Go module with the following key characteristics:
- Implements a TCP server that handles Redis protocol (RESP - Redis Serialization Protocol)
- Supports configuration via a redis.conf file
- Features a modular architecture with separate components for networking, parsing, and data handling
- Includes graceful shutdown capabilities

## Architecture

The project follows a modular architecture with these main components:

### Core Components

1. **Main Entry Point** (`main.go`): Initializes configuration and starts the TCP server
2. **Configuration** (`config/config.go`): Handles loading and parsing of server configuration from redis.conf
3. **TCP Server** (`tcp/`): Implements the TCP server with graceful shutdown capabilities
4. **RESP Handler** (`resp/`): Handles the Redis Serialization Protocol for command parsing and response
5. **Database Layer** (`database/`): Implements the data storage and command execution logic
6. **Connection Management** (`resp/connection/`): Manages client connections
7. **Parser** (`resp/parser/`): Parses incoming RESP protocol commands
8. **Reply System** (`resp/reply/`): Formats responses according to RESP protocol

### Configuration

The server supports configuration through a `redis.conf` file with the following parameters:
- `bind`: IP address to bind to (default: 0.0.0.0)
- `port`: Port number to listen on (default: 6379)
- `appendOnly`: Enable append-only file persistence
- `appendFilename`: Name of the append-only file
- `maxClients`: Maximum number of concurrent clients
- `requirePass`: Password for authentication
- `databases`: Number of databases
- `peers`: List of peer servers for clustering
- `self`: Self identifier for clustering

## Building and Running

### Prerequisites
- Go 1.23.1 or higher

### Building
The project can be built using standard Go commands:
```bash
go build -o redis-go main.go
```

### Running
To run the server:
```bash
go run main.go
```

The server will read configuration from `redis.conf` in the current directory and start listening on the configured address and port.

### Testing
The project doesn't appear to have explicit test files in the current structure, but can be tested by connecting to the server with a Redis client.

## Development Conventions

### Code Structure
- The code follows Go best practices with clear package separation
- Each major component is in its own directory with related functionality
- Interfaces are defined in the `interface/` directory to promote loose coupling
- Error handling is consistent throughout the codebase

### Naming Conventions
- Public functions and types use PascalCase
- Private functions and types use camelCase
- Configuration tags use the `cfg` tag for struct fields

### Design Patterns
- The project uses the Handler pattern for processing client connections
- Connection pooling and management through sync.Map
- Graceful shutdown using signal handling
- Context propagation for request lifecycle management

## Key Features

1. **TCP Server with Signal Handling**: The server can handle multiple concurrent connections and responds to system signals for graceful shutdown.

2. **RESP Protocol Support**: Implements the Redis Serialization Protocol for command parsing and response formatting.

3. **Configuration Management**: Supports loading server configuration from a file with various parameters.

4. **Connection Management**: Tracks active connections and handles client lifecycle events.

5. **Graceful Shutdown**: Properly closes connections and resources when shutting down.

## Project Status

This is a learning-focused implementation that currently appears to be in a basic working state. The database implementation is currently an "echo" database that simply returns the command arguments, suggesting this is a foundation for building more complex Redis command implementations.

## Dependencies

The project uses only standard Go libraries with no external dependencies declared in the go.mod file, making it lightweight and self-contained.