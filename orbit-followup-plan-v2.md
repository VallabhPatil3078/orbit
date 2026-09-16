# Orbit — Follow-Up Plan v2

Round 1 is mostly done. This round has two parts: **(A)** one real bug from the last review plus a few remaining polish items, and **(B)** your GitHub Actions setup, broken into concrete steps since that's the professionalism signal a recruiter/interviewer will actually click on.

---

## A. Code fixes

### A1. Fix the non-portable concurrency test (blocking — this will fail on Linux/macOS CI)

**The problem:** `TestExecuteTiers_RunsConcurrently` hardcodes `powershell -c "Start-Sleep 1"`. Your `runner.go` correctly branches on `runtime.GOOS`, but this test doesn't — so on a Linux GitHub Actions runner (the default), `sh -c "powershell -c ..."` fails because PowerShell doesn't exist there, and the test goes red on a perfectly correct codebase. **Fix this before wiring up CI, or your very first Actions run will fail.**

**Questions to answer:**
- Do you want the test to shell out to a real OS sleep command (branched by `runtime.GOOS`, same pattern as `runner.go`), or avoid shelling out entirely and test concurrency more directly (e.g. each task increments a shared counter with a small delay, no `exec.Command` involved)?

**Fix to implement (`pkg/runner/runner_test.go`):**
- [ ] Replace the hardcoded PowerShell command with a `runtime.GOOS`-branched sleep, matching the pattern already in `runner.go`:
  ```go
  var sleepCmd string
  if runtime.GOOS == "windows" {
      sleepCmd = "powershell -c \"Start-Sleep 1\""
  } else {
      sleepCmd = "sleep 1"
  }
  ```
- [ ] Re-run the test locally (or in CI once A2 is done) to confirm it passes on your actual runner OS, not just on your dev machine.

### A2. Add the missing malformed-YAML test (small, non-blocking)

**Fix to implement (`pkg/config/parser_test.go`):**
- [ ] Add `TestParseConfig_MalformedYAML`: write a temp file with broken YAML syntax (e.g. unclosed bracket or bad indentation), call `ParseConfig`, assert a non-nil error. The code already handles this (you propagate `yaml.Unmarshal`'s error) — this just closes the test gap.

### A3. Optional polish (only if you have spare time before moving on)
- [ ] `README.md` — confirm it documents `orbit validate` and the `--force` flag on `init`, since both were added after the original walkthrough was written.
- [ ] Consider whether `ParseConfig` should also reject a config with zero tasks defined — currently an empty `tasks:` map parses successfully and produces a 0-tier no-op run. Decide if that's intentional (harmless no-op) or should be a validation error.

---

## B. GitHub Actions CI setup (your task)

Goal: on every push/PR, GitHub automatically builds, vets, and tests Orbit — the standard signal that a Go repo is taken seriously.

### B1. Decide the trigger and matrix
**Questions to answer:**
- Run on every push to `main` only, or also on every pull request? (Recommendation: both — PRs are where CI actually catches things before they land.)
- Test on Linux only, or a matrix of Linux + Windows + macOS? Given Orbit's whole pitch is cross-platform `exec` handling, testing on **at least Linux and Windows** is the strongest proof that claim is real — and it's exactly the test suite (after A1's fix) that would catch a regression there.

### B2. Write the workflow file
**Fix to implement (`.github/workflows/ci.yml`):**
- [ ] Create the workflow with:
  - `on: push` (branches: `main`) and `on: pull_request`
  - A matrix job: `os: [ubuntu-latest, windows-latest]`
  - Steps: checkout → set up Go (pin the version, e.g. `1.22.x`, don't leave it floating) → `go build ./...` → `go vet ./...` → `go test ./... -v`
- [ ] Confirm `go.mod`'s `go` directive version is compatible with whatever Go version you pin in the workflow (mismatches here are a common first-CI-run failure).

### B3. Add a status badge
- [ ] Add the workflow status badge to the top of `README.md` (`![CI](https://github.com/<you>/<repo>/actions/workflows/ci.yml/badge.svg)`) — this is the visual "this project is tested" signal recruiters actually notice.

### B4. Branch protection (optional but professional)
**Questions to answer:**
- Do you want `main` to require the CI check to pass before merge? If this is a solo repo, this mostly matters as a demonstrated habit (interviewers sometimes check repo settings), not as an enforcement mechanism you personally need.

**Fix to implement (GitHub repo Settings, not code):**
- [ ] Settings → Branches → Add rule for `main` → require status checks to pass before merging → select the CI workflow.

### B5. Sanity check before you call it done
- [ ] Push a small throwaway commit (or open a PR) after setting this up and confirm the Actions tab actually goes green — specifically confirm the Windows matrix leg passes, since that's the leg that validates your `cmd /C` branch, which nothing else in your test suite currently exercises on a real Windows runner.

---

## Suggested order
1. **A1 first** — fix the test before you wire up CI, or your first Actions run fails for a reason that has nothing to do with your CI config, and you'll waste time debugging the wrong thing.
2. **B1–B3** — get the workflow running and green.
3. **A2** — quick test gap, fits in CI once it exists.
4. **B4–B5** — protection rule + final green-run confirmation.
5. **A3** — only if time remains.
