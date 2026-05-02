#!/bin/bash

export PORT=${1:-8080}

echo "Starting Stock Market Simulator on localhost:${PORT}..."

docker compose down
docker compose up --build -d

echo "Application is running at http://localhost:${PORT}"