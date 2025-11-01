#!/bin/bash
# Wrapper script to build GPU Pro using Docker
# Works on any platform (Linux, macOS, Windows with WSL)

set -e

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  GPU Pro - Docker Build${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}Error: Docker is not installed${NC}"
    echo "Please install Docker from: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    echo -e "${YELLOW}Error: Docker daemon is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi

echo -e "${GREEN}✓${NC} Docker is ready"
echo

# Build the builder image
echo -e "${BLUE}Building Docker image...${NC}"
docker-compose build builder

echo
echo -e "${BLUE}Starting cross-platform build...${NC}"
echo

# Run the build
docker-compose run --rm builder

echo
echo -e "${GREEN}✓ Build complete!${NC}"
echo
echo -e "${BLUE}Built binaries are in:${NC} ./dist/"
echo
echo -e "${BLUE}To clean up Docker resources:${NC}"
echo "  docker-compose down --volumes"
echo
