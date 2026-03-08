#!/bin/bash

# Simplified test helper for containerized testing
# In containers, we don't need the complex local/container detection

# Set up containerized test environment
export TTR_BASE_DIR="/home/testuser/.timetracker_test"
export HOME="/home/testuser"
export TEST_ENVIRONMENT="containerized"

# Create necessary directories
mkdir -p "$TTR_BASE_DIR/records"
mkdir -p "$TTR_BASE_DIR/sessions"

echo "✅ Containerized test environment ready"