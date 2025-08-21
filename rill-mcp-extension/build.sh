#!/bin/bash

# Rill MCP Desktop Extension Build Script

set -e

echo "🔧 Building Rill MCP Desktop Extension..."

# Check if we're in the right directory
if [ ! -f "manifest.json" ]; then
    echo "❌ Error: manifest.json not found. Run this script from the extension directory."
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
npm install --production

# Validate and build extension
echo "📦 Packaging extension..."
dxt pack

# Get the package name and version from manifest
PACKAGE_NAME=$(node -p "require('./manifest.json').name")
PACKAGE_VERSION=$(node -p "require('./manifest.json').version")

echo "✨ Successfully created: ${PACKAGE_NAME}-${PACKAGE_VERSION}.dxt"
echo ""
echo "🎉 Build complete!"
echo ""
echo "To install:"
echo "1. Open Claude Desktop"
echo "2. Go to Settings → Extensions" 
echo "3. Drag and drop ${PACKAGE_NAME}-${PACKAGE_VERSION}.dxt"
echo "4. Configure your Rill project URL and token"
echo ""
echo "Documentation: https://docs.rilldata.com/explore/mcp"
