# Stock-Sim

A highly available, simplified stock exchange service written in Go.

## Prerequisites

* Docker
* Docker Compose
* Docker daemon must be running

## Running the Application

You can start the application with a single command.
If no port is specified, the service runs on **port 8080** by default.

---

### Linux / macOS

```bash
# Make the script executable (only needed once)
chmod +x start.sh

# Run on default port (8080)
./start.sh

# Run on a custom port (e.g., 3000)
./start.sh 3000
```

---

### Windows

```bat
:: Run on default port (8080)
start.bat

:: Run on a custom port (e.g., 3000)
start.bat 3000
```


## Tests

### Unit Tests

```bash
go test ./internal/service ./internal/handler -v
```
