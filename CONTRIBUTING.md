# Contributing to Orbit

First off, thank you for considering contributing to Orbit! 

To keep the repository clean and maintainable, please follow these guidelines when contributing:

## Branching Strategy
- Do not push directly to `main`.
- Create a feature branch for your work: `git checkout -b feature/your-feature-name` or `git checkout -b fix/your-bug-fix`.

## Local Development & Testing
Before opening a Pull Request, you must ensure that your code meets the quality standards:
1. **Format your code**: Run `go fmt ./...` to ensure consistent code styling.
2. **Run the tests**: Run `go test ./...` and ensure all tests pass.
3. **Add tests**: Any new feature or bug fix must include corresponding tests.

## Commit Messages
We follow [Conventional Commits](https://www.conventionalcommits.org/). Please format your commit messages clearly:
- `feat: added a new feature`
- `fix: resolved a bug`
- `docs: updated documentation`
- `chore: routine tasks`

## Pull Requests
- Provide a clear and descriptive title for your PR.
- Fill out the PR description with the problem you are solving and how you solved it.
- Ensure the CI workflow passes.
