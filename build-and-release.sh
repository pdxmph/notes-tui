#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from argument or prompt
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo -n "Enter version (e.g., v0.1.0): "
    read VERSION
fi

# Validate version format
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Version must be in format vX.Y.Z${NC}"
    exit 1
fi

echo -e "${GREEN}🚀 Starting release process for $VERSION${NC}"

# Step 1: Ensure all changes are committed
echo -e "\n${YELLOW}Step 1: Checking for uncommitted changes...${NC}"
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo -e "${RED}Error: Uncommitted changes found. Please commit or stash them first.${NC}"
    exit 1
fi

# Step 2: Create and push tag
echo -e "\n${YELLOW}Step 2: Creating tag...${NC}"
if git tag | grep -q "^${VERSION}$"; then
    echo -e "${RED}Error: Tag $VERSION already exists${NC}"
    exit 1
else
    echo "Enter release notes (press Ctrl-D when done):"
    RELEASE_NOTES=$(cat)
    git tag -a "$VERSION" -m "$RELEASE_NOTES"
fi

# Step 3: Push tag to GitHub
echo -e "\n${YELLOW}Step 3: Pushing tag to GitHub...${NC}"
git push origin "$VERSION"

# Step 4: Build binaries for current platform
echo -e "\n${YELLOW}Step 4: Building binaries...${NC}"
rm -rf dist
mkdir -p dist

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

# Build the binary
echo "Building for $OS/$ARCH..."
go build -o "dist/notes-tui" .

# Create archive
ARCHIVE_NAME="notes-tui_${VERSION}_${OS}_${ARCH}.tar.gz"
cd dist
cp ../README.md ../config.example.toml .
FILES_TO_ARCHIVE="notes-tui README.md config.example.toml"
if [ -f ../LICENSE ]; then
    cp ../LICENSE .
    FILES_TO_ARCHIVE="$FILES_TO_ARCHIVE LICENSE"
fi
tar czf "$ARCHIVE_NAME" $FILES_TO_ARCHIVE
rm README.md config.example.toml
if [ -f LICENSE ]; then
    rm LICENSE
fi
cd ..

# Create checksums
cd dist
shasum -a 256 *.tar.gz > checksums.txt
cd ..

# Step 5: Create GitHub release
echo -e "\n${YELLOW}Step 5: Creating GitHub release...${NC}"
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) not found. Install with: brew install gh${NC}"
    echo -e "${YELLOW}Alternatively, create the release manually at:${NC}"
    echo "https://github.com/pdxmph/notes-tui/releases/new?tag=$VERSION"
    echo "Upload these files:"
    ls -la dist/*.tar.gz dist/checksums.txt
else
    gh release create "$VERSION" \
      --title "$VERSION" \
      --notes "$RELEASE_NOTES" \
      dist/*.tar.gz \
      dist/checksums.txt
fi

echo -e "\n${GREEN}✅ Release $VERSION completed successfully!${NC}"
echo -e "${GREEN}View at: https://github.com/pdxmph/notes-tui/releases/tag/$VERSION${NC}"

# Show installation instructions
echo -e "\n${YELLOW}Installation instructions:${NC}"
echo "Download and install:"
echo "  curl -L https://github.com/pdxmph/notes-tui/releases/download/$VERSION/$ARCHIVE_NAME | tar xz"
echo "  sudo mv notes-tui /usr/local/bin/"
echo ""
echo "Or download directly from: https://github.com/pdxmph/notes-tui/releases/tag/$VERSION"
