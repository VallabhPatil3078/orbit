# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Config Schema versioning (implicitly `version: 1`) to ensure forwards-compatibility.
- Per-task timeout configuration in `orbit.yaml`.
- OS-level process tree cleanup (Job Objects on Windows, Process Groups on Unix) for proper cancellation.
- Graceful shutdown sequence (`SIGTERM` -> 5s -> `SIGKILL`) on Unix platforms.
- `orbit version` command injected at build time.

### Changed
- Standardized exit codes (1: Config Error, 2: DAG Error, 3: Task Failure, 4: Timeout).
- Concurrent execution model now bounds all tasks with explicit context timeouts.
- Quiet mode (`--quiet`) for minimal CI logging.
- Module path renamed to `github.com/VallabhPatil3078/orbit`.

### Fixed
- Fixed cascade skip evaluation logic to correctly evaluate OR filters.
- Re-normalized line-endings to `LF` to fix cross-platform hash inconsistencies.
- Eliminated Windows CI flakes by removing `powershell` dependencies from tests.
