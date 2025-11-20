# Architecture Graphs

This directory contains automatically generated architecture graphs using `go-arch-lint`.

## Directory Structure

- `flow/` - Flow graphs showing reverse dependency injection (default)
- `dependency-injection/` - DI graphs showing component dependencies  
- `focused/` - Component-specific focused graphs
- `vendor/` - Graphs including vendor dependencies

## Generation Commands

See the justfile for available graph commands:
- `just graph` - Generate main flow graph
- `just graph-all` - Generate ALL graph types
- `just graph-di` - Generate dependency injection graph
- `just graph-vendor` - Generate graph with vendors
- `just graph-component <name>` - Generate focused component graph

## Understanding the Graphs

- **Flow graphs** (default): Show execution flow (reverse dependency injection)
- **DI graphs**: Show direct component dependencies
- **Vendor graphs**: Include external library dependencies
- **Focused graphs**: Show single component and its dependencies

Generated: Thu Nov 20 22:04:03 CET 2025

