#!/bin/bash

# Master script to compile all proto files for the DVPN project
# This script compiles proto files for both clientPeer and super directories

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== DVPN Proto Compilation Master Script ==="
echo

# Function to compile proto files for a given directory
compile_proto_for_dir() {
    local dir_name=$1
    local dir_path="$SCRIPT_DIR/$dir_name"

    if [ ! -d "$dir_path" ]; then
        echo "❌ Directory $dir_name not found, skipping..."
        return 1
    fi

    if [ ! -f "$dir_path/compile_proto.sh" ]; then
        echo "❌ compile_proto.sh not found in $dir_name, skipping..."
        return 1
    fi

    echo "🔧 Compiling proto files for $dir_name..."
    cd "$dir_path"
    ./compile_proto.sh
    echo "✅ $dir_name compilation completed"
    echo

    cd "$SCRIPT_DIR"
    return 0
}

# Compile for clientPeer
echo "📁 Processing clientPeer directory..."
compile_proto_for_dir "clientPeer"

# Compile for super
echo "📁 Processing super directory..."
compile_proto_for_dir "super"

# Test builds
echo "🧪 Testing builds..."
echo

echo "Testing clientPeer build..."
cd "$SCRIPT_DIR/clientPeer"
if go build; then
    echo "✅ clientPeer builds successfully"
else
    echo "❌ clientPeer build failed"
    exit 1
fi
echo

echo "Testing super build..."
cd "$SCRIPT_DIR/super"
if go build; then
    echo "✅ super builds successfully"
else
    echo "❌ super build failed"
    exit 1
fi

cd "$SCRIPT_DIR"
echo
echo "🎉 All proto files compiled and builds tested successfully!"
echo
echo "Generated files:"
echo "📁 clientPeer/pb/:"
ls -1 clientPeer/pb/*.pb.go | sed 's/^/  /'
echo
echo "📁 super/pb/:"
ls -1 super/pb/*.pb.go | sed 's/^/  /'
echo
echo "To compile proto files individually:"
echo "  ./clientPeer/compile_proto.sh"
echo "  ./super/compile_proto.sh"
