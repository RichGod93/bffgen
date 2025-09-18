#!/bin/bash

# bffgen Release Script
# Usage: ./scripts/release.sh v0.1.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ -z "$1" ]; then
    print_error "Version is required!"
    echo "Usage: $0 v0.1.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Invalid version format: $VERSION"
    echo "Expected format: v0.1.0"
    exit 1
fi

print_status "Starting release process for $VERSION"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    print_warning "Not on main branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Release cancelled"
        exit 1
    fi
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    print_error "Working directory is not clean!"
    git status --short
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    print_error "Tag $VERSION already exists!"
    exit 1
fi

print_status "Running tests..."
if ! make test; then
    print_error "Tests failed!"
    exit 1
fi

print_status "Running linter..."
if ! make lint; then
    print_error "Linting failed!"
    exit 1
fi

print_status "Building for all platforms..."
if ! make build-all VERSION="$VERSION"; then
    print_error "Build failed!"
    exit 1
fi

print_status "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

print_status "Pushing tag to remote..."
git push origin "$VERSION"

print_success "Release $VERSION created successfully!"
print_status "GitHub Actions will now build and publish the release"
print_status "View release: https://github.com/RichGod93/bffgen/releases/tag/$VERSION"

# Optional: Create a release branch for hotfixes
read -p "Create release branch 'release/$VERSION' for hotfixes? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git checkout -b "release/$VERSION"
    git push origin "release/$VERSION"
    git checkout main
    print_success "Release branch 'release/$VERSION' created"
fi

print_success "Release process completed!"
