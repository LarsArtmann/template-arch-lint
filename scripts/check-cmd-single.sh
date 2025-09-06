#!/bin/bash

# Script to enforce single main.go file in cmd/ directory
# Used by justfile lint-cmd-single command
#
# This script implements architectural constraint enforcement ensuring
# exactly one main.go file exists in the cmd/ directory, promoting
# clean architecture patterns and preventing command proliferation.

set -euo pipefail

echo -e "\033[1m🎯 CMD SINGLE MAIN ENFORCEMENT\033[0m"
echo -e "\033[0;36mChecking for single main.go file in cmd/ directory...\033[0m"

# Check if cmd/ directory exists and is accessible
if [ ! -d "cmd" ]; then
    echo -e "\033[0;31m❌ cmd/ directory not found!\033[0m"
    echo -e "\033[0;33m💡 Create cmd/ directory with a single main.go file for your application entry point.\033[0m"
    exit 1
fi

# Check if cmd/ directory is readable
if [ ! -r "cmd" ]; then
    echo -e "\033[0;31m❌ cmd/ directory is not readable! Permission denied.\033[0m"
    echo -e "\033[0;33m💡 Check file permissions: chmod 755 cmd\033[0m"
    exit 1
fi

# Find main.go files with robust error handling
# Use -L to follow symbolic links, but also detect broken links
main_files_output=""
find_exit_code=0

# First pass: find all main.go files (following symlinks)
if ! main_files_output=$(find cmd -name "main.go" -type f -o -name "main.go" -type l 2>/dev/null); then
    find_exit_code=$?
    echo -e "\033[0;31m❌ Error searching cmd/ directory (exit code: $find_exit_code)\033[0m"
    exit 1
fi

# Count main.go files more robustly
main_count=0
if [ -n "$main_files_output" ]; then
    # Use proper counting that handles edge cases
    main_count=$(echo "$main_files_output" | grep -c '^' || echo 0)
fi

# Validate the constraint and provide actionable feedback
if [ "$main_count" -eq 0 ]; then
    echo -e "\033[0;31m❌ No main.go files found in cmd/ directory!\033[0m"
    echo -e "\033[0;31m   Expected exactly 1 main.go file for clean architecture.\033[0m"
    echo -e "\033[0;33m💡 SOLUTION: Create a main.go file in a subdirectory:\033[0m"
    echo -e "\033[0;33m   mkdir -p cmd/server && echo 'package main\n\nfunc main() {\n    // Your app logic\n}' > cmd/server/main.go\033[0m"
    exit 1
elif [ "$main_count" -gt 1 ]; then
    echo -e "\033[0;31m❌ Found $main_count main.go files in cmd/ directory:\033[0m"
    echo "$main_files_output" | sed 's/^/   ❌ /'
    echo -e "\033[0;31m   Expected exactly 1 main.go file for clean architecture.\033[0m"
    echo ""
    echo -e "\033[0;33m💡 SOLUTION: Consolidate multiple commands into subcommands:\033[0m"
    echo -e "\033[0;33m   1. Use a CLI framework like cobra or urfave/cli\033[0m"
    echo -e "\033[0;33m   2. Create a single main.go with subcommands:\033[0m"
    echo -e "\033[0;33m      cmd/server/main.go with 'server start', 'server migrate', etc.\033[0m"
    echo -e "\033[0;33m   3. Or move additional commands to separate packages/tools\033[0m"
    echo ""
    echo -e "\033[0;33m📚 Learn more: https://pkg.go.dev/github.com/spf13/cobra\033[0m"
    exit 1
else
    echo -e "\033[0;32m✅ Found exactly 1 main.go file in cmd/ - architecture constraint satisfied!\033[0m"
    echo "$main_files_output" | sed 's/^/   📁 /'
    echo -e "\033[0;32m   Clean architecture pattern maintained! ✨\033[0m"
fi

# Additional validation: check if main.go is actually a valid Go main package
if [ "$main_count" -eq 1 ]; then
    main_file_path="$main_files_output"
    if [ -r "$main_file_path" ]; then
        # Basic validation that it's a main package
        if ! grep -q "^package main" "$main_file_path" 2>/dev/null; then
            echo -e "\033[0;33m⚠️  WARNING: $main_file_path does not appear to be a 'package main' file\033[0m"
            echo -e "\033[0;33m   Ensure it starts with 'package main' for executable programs\033[0m"
        fi
        
        # Check for main function
        if ! grep -q "func main()" "$main_file_path" 2>/dev/null; then
            echo -e "\033[0;33m⚠️  WARNING: $main_file_path does not contain a 'func main()' function\033[0m"
            echo -e "\033[0;33m   Add a main() function as the program entry point\033[0m"
        fi
    fi
fi