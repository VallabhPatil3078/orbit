# Orbit 🪐

**Orbit** is a localized, high-performance DAG-based task orchestrator designed to act as an instant CI/CD gatekeeper for your Git commits. 

By intercepting Git hooks (like `pre-commit`), Orbit intelligently schedules validation checks—such as formatters, linters, and unit tests. Instead of naive parallel execution, Orbit parses task configurations into a Directed Acyclic Graph (DAG), maximizing CPU efficiency by executing independent tasks concurrently while strictly enforcing execution order for dependent tasks.

## Features
- **DAG Engine:** Parses tasks using Topological Sort (Kahn's algorithm).
- **Maximum Concurrency:** Independent tasks run in parallel using Go Goroutines.
- **Zero-Dependency:** A single Go binary that doesn't bloat your project repository.
- **Git Hook Integration:** Automatically intercepts commits to prevent broken code from being pushed.

## How it works
Orbit looks for an `orbit.yaml` file in the root of your project:

```yaml
tasks:
  lint:
    command: "npm run lint"
    depends_on: []
  format:
    command: "prettier --write ."
    depends_on: []
  build:
    command: "npm run build"
    depends_on: ["lint", "format"] # Build waits until lint and format finish
```

*(This project is currently under active development).*
