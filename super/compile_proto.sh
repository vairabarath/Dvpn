#!/bin/bash

# Script to compile proto files for the DVPN super node

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROTO_DIR="$SCRIPT_DIR/proto"
OUTPUT_DIR="$SCRIPT_DIR/pb"

echo "Compiling proto files for super node..."

# Remove old generated files
echo "Removing old generated files..."
rm -f "$OUTPUT_DIR"/*.pb.go

# Compile all proto files
echo "Generating new proto files..."
protoc --proto_path="$PROTO_DIR" \
       --go_out="$OUTPUT_DIR" \
       --go-grpc_out="$OUTPUT_DIR" \
       "$PROTO_DIR"/*.proto

# Check if files were generated in nested directories due to go_package option
if [ -d "$OUTPUT_DIR/pb" ]; then
    echo "Moving files from nested directory..."
    mv "$OUTPUT_DIR/pb"/*.pb.go "$OUTPUT_DIR/"
    rm -rf "$OUTPUT_DIR/pb"
fi

echo "Proto compilation completed successfully!"
echo "Generated files:"
ls -la "$OUTPUT_DIR"/*.pb.go
