#!/bin/bash

# Script to enforce single main.go file in cmd/ directory
# Used by justfile lint-cmd-single command

set -euo pipefail

echo -e "\033[1m🎯 CMD SINGLE MAIN ENFORCEMENT\033[0m"
echo -e "\033[0;36mChecking for single main.go file in cmd/ directory...\033[0m"

# Check if cmd/ directory exists
if [ ! -d "cmd" ]; then
    echo -e "\033[0;31m❌ cmd/ directory not found!\033[0m"
    exit 1
fi

# Count main.go files in cmd/ directory
main_files=$(find cmd -name "main.go" -type f)
if [ -z "$main_files" ]; then
    main_count=0
else
    main_count=$(echo "$main_files" | wc -l | tr -d ' ')
fi

if [ "$main_count" -eq 0 ]; then
    echo -e "\033[0;31m❌ No main.go files found in cmd/ directory!\033[0m"
    echo -e "\033[0;31m   Expected exactly 1 main.go file for clean architecture.\033[0m"
    exit 1
elif [ "$main_count" -gt 1 ]; then
    echo -e "\033[0;31m❌ Found $main_count main.go files in cmd/ directory:\033[0m"
    echo "$main_files" | sed 's/^/   /'
    echo -e "\033[0;31m   Expected exactly 1 main.go file for clean architecture.\033[0m"
    echo -e "\033[0;33m💡 Consider consolidating multiple commands into a single main with subcommands.\033[0m"
    exit 1
else
    echo -e "\033[0;32m✅ Found exactly 1 main.go file in cmd/ - architecture constraint satisfied!\033[0m"
    echo "$main_files" | sed 's/^/   📁 /'
fi