# Project Split Executive Report: template-arch-lint

## Introduction

The `template-arch-lint` project, as described in its `AGENTS.md`, currently serves multiple distinct purposes: a Go linting template, a Clean Architecture and DDD reference implementation, a configuration library, and a simple HTMX demo web application. While this comprehensive approach is valuable as an educational resource and template, splitting it into more focused projects would enhance maintainability, reusability, and clarity of purpose for each component.

This report proposes a logical decomposition of the existing monolithic project into several highly focused, independent repositories.

## Proposed Project Structure

Based on the core functionalities and stated purpose of `template-arch-lint`, the following distinct projects are proposed:

### 1. `go-lint-configs`

- **Purpose**: A dedicated repository for opinionated, enterprise-grade linting configurations and tooling for Go projects. This project would serve as a standalone resource for teams to adopt robust code quality and architectural enforcement.
- **Key Components**:
  - `.go-arch-lint.yml`: Architectural boundary rules.
  - `.golangci.yml`: Comprehensive Go linter configurations (99+ linters).
  - `justfile`: Task automation scripts specifically for linting, fixing, and security auditing (e.g., `just lint`, `just fix`, `just security-audit`).
  - `README.md`: Detailed instructions on integrating these configurations into new or existing Go projects.
- **Rationale for Split**: This component is explicitly described as a "Configuration Library" meant for copy/paste. Separating it allows direct consumption without the overhead of the reference application.

### 2. `go-clean-arch-ddd-reference`

- **Purpose**: A focused reference implementation of Clean Architecture and Domain-Driven Design patterns in Go. This project would illustrate best practices for structuring Go applications with clear layer boundaries, rich domain models, value objects, and robust error handling using a simple, illustrative business domain (e.g., User CRUD).
- **Key Components**:
  - `internal/domain/`: Pure business logic (entities, services, repositories interfaces, value objects, domain errors, result pattern).
  - `internal/application/`: Application-specific logic (e.g., use cases, DTOs, HTTP response helpers, middleware _if truly application-generic_).
  - `internal/infrastructure/`: External concerns (persistence implementations, database-specific code).
  - `internal/db/`: SQLC-generated type-safe SQL code (schema, queries).
  - Minimal `cmd/server/main.go` demonstrating the application's entry point and dependency injection.
  - BDD-style tests (`Ginkgo/Gomega`) for domain and application logic.
- **Rationale for Split**: This is the "Reference Implementation" aspect. Decoupling it from the specific linting configs and the demo web app allows it to be a pure architectural example, easier to study and adapt.

### 3. `go-htmx-templ-starter`

- **Purpose**: A lightweight, opinionated starter project for building modern web applications using Go for the backend, HTMX for progressive enhancement, and Templ for type-safe server-side rendering. This project would focus on the integration of these web technologies with a clear, minimal structure.
- **Key Components**:
  - `web/templates/`: Templ components, layouts, and pages.
  - `internal/application/handlers/`: HTTP handlers specifically for serving web content and handling HTMX requests.
  - `internal/application/dto/`: Data transfer objects for web interactions.
  - `internal/application/http/`: HTTP response helpers tailored for web contexts.
  - `cmd/server/main.go`: The primary entry point for the web server, demonstrating setup and routing.
- **Rationale for Split**: The current project includes a "Simple HTMX Demo". By separating this into its own starter, it becomes a directly usable template for anyone wanting to build Go+HTMX+Templ applications, without needing the full Clean Architecture reference or the linting configurations. This project would likely depend on the `go-clean-arch-ddd-reference` (or a simplified version of its domain layer) for business logic.

### 4. `go-arch-graph-tools` (Optional Library/CLI)

- **Purpose**: A reusable Go library or CLI tool for statically analyzing Go projects and generating visual architecture graphs (flow, dependency injection, vendor, focused component graphs). This would abstract the graph generation logic into a standalone, extensible tool.
- **Key Components**:
  - Go packages for parsing Go source code, analyzing dependencies, and generating graph data.
  - CLI entry point (e.g., `arch-graph-cli`) to invoke graph generation.
  - Output formats (SVG, DOT).
  - Integration points for `go-arch-lint` or similar tools.
- **Rationale for Split**: The architecture graphing capability is a distinct tool concern. Extracting it as a separate library or CLI tool makes it independently usable and distributable, potentially for other Go projects.

## Conclusion

This proposed split transforms the `template-arch-lint` project from a multi-purpose template into a suite of highly specialized and reusable components. Each new project addresses a specific problem domain, offering greater clarity, reducing cognitive load for adopters, and improving the overall modularity and maintainability of the codebase. This approach aligns with the principles of single responsibility and promotes easier integration into diverse Go development workflows.
