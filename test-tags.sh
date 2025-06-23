#!/bin/bash
# Test script for tag prompting feature

# Create test directory if it doesn't exist
mkdir -p test-notes

# Run notes-tui with test config
./notes-tui -config test-tag-config.toml

# After testing, show any created notes
echo "Notes created during testing:"
ls -la test-notes/
