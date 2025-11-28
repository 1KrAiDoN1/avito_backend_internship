#!/bin/bash
set -e

echo "Running migrations..."
migrate -path ./migrations -database "${DB_URL}" up



