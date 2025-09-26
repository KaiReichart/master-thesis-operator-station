#!/bin/bash
# Cleanup old template files after migration to templ

echo "Cleaning up old template files..."

# Remove old template files (they're already backed up as .old)
rm -f programs/templates.go.old
rm -f gps/templates.go.old  
rm -f events/templates.go.old

echo "Old template files removed."
echo ""
echo "To generate templ files, run: make generate"
echo "To build the project, run: make build"
echo "To run in development mode, run: make run"
