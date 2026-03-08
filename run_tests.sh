#!/bin/bash

# TimeTracker Safe Test Runner
# Containerized testing to protect your local environment

echo "🚀 TimeTracker Test Runner"
echo "=========================="
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker to run tests safely."
    echo "   Visit https://docs.docker.com/get-docker/ for installation instructions."
    exit 1
fi

# Check Docker is running
if ! docker info &> /dev/null; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

echo "✅ Docker is available and running"
echo ""

# Build the test container
echo "🐳 Building test container..."
if ! docker build -t timetracker-tests -f Dockerfile.test .; then
    echo "❌ Failed to build test container"
    exit 1
fi

echo "✅ Test container built successfully"
echo ""

# Run the tests in the container
echo "🧪 Running tests in isolated container..."
echo ""

if docker run --rm timetracker-tests; then
    echo ""
    echo "🎉 All tests completed successfully!"
    echo "🧹 Container automatically cleaned up"
    exit 0
else
    echo ""
    echo "❌ Some tests failed"
    echo "🧹 Container automatically cleaned up"
    exit 1
fi