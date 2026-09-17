# Orbit — Follow-Up Plan v4

Two parts: **(A)** the two bugs from the last review (one blocking, one a correctness inversion), and **(B)** a judged review of your CLI logging revamp proposal (`implementation_plan.md`) with the gaps it should close before you implement it.

---

## A. Outstanding bugs from the trigger_paths/ignore_paths review

### A1. No changed files ≠ skip everything (blocking — breaks the shipped demo config)

**The problem:** in `runner.go`, when `changedFiles` is empty, the `for _, file := range changedFiles` loop never runs, so `shouldSkip` stays at its initial `true` for every task with any `trigger_paths`/`ignore_paths`. There's no fallback for "nothing staged." Since every task in the shipped `orbit.yaml` has `trigger_paths` set, a fresh clone + `orbit run` with nothing staged currently skips everything and exits "successfully" having done nothing.

**Fix to implement (`pkg/runner/runner.go`):**
- [x] When `changedFiles` is empty, `baseRef == ""`, and `!forceAll`: either default to running everything, or print an explicit message (`[!] No staged changes detected — did you mean --all?`) instead of silently skipping every task.
- [x] Add a test: `ExecuteTiers` called with `changedFiles: []`, `forceAll: false` — assert the chosen behavior (either all tasks run, or a clear warning is emitted) rather than the current silent full-skip.

### A2. Cascade-skip logic uses OR when it should use AND

**The problem:** a child task with no `trigger_paths`/`ignore_paths` of its own currently skips if **any** dependency was skipped:
```go
for _, dep := range n.DependsOn {
    if skipStates[dep] { shouldSkip = true; break }
}
```
This is backwards for a child with multiple parents that have different trigger criteria — it should skip only if **all** its parents were skipped (nothing upstream happened at all), and run if at least one parent actually ran. As written, one irrelevant skipped sibling can suppress a task that should have run.

**Fix to implement (`pkg/runner/runner.go`):**
- [x] Flip the condition: skip only when every dependency in `n.DependsOn` has `skipStates[dep] == true`.
- [x] Add a test with two parents that have *different* trigger paths (e.g. parent X triggers on `**/*.go`, parent Y triggers on `**/*.md`) and a child with no filters depending on both — confirm the child runs when only X runs, and only skips when both X and Y skip. This is the scenario your current `skip_test.go` doesn't cover (it only tests a single-parent case), which is why the bug shipped unnoticed.

### A3. Smaller items
- [x] `doublestar.Match`'s error is discarded (`match, _ := doublestar.Match(...)`). A malformed glob silently never matches, with no indication why a task keeps skipping. Surface this in `orbit validate` — at minimum, call `doublestar.Match` on a throwaway string for each configured pattern at validate-time and report a bad pattern explicitly.
- [x] Add a test with both `trigger_paths` and `ignore_paths` set on the same task (not yet covered) — confirm ignore excludes a file first, then trigger is evaluated against what's left.

---

## B. Judging the CLI logging revamp (`implementation_plan.md`)

**The core idea is good** — hiding "Tier" as an implementation detail and moving to a cleaner `✔/✖/-` format is the right instinct, and it's genuinely closer to how Vite/Turborepo present output. But the plan as written has one omission that will break your test suite silently, plus two issues worth deciding before you write code.

### B1. This will break `skip_test.go` and nobody will notice until CI fails (blocking — sequence this first)

**The problem:** `skip_test.go` verifies skip/run behavior by pattern-matching the *exact printed strings* Orbit emits today — `"[-] Task '" + task + "' skipped"` and `"[V] Task '" + task + "' finished"`. The implementation plan changes these to `"  - %s (skipped: ...)"` and `"  ✔ %s (completed)"` but never mentions updating the tests. If you implement the plan as written, `skip_test.go` goes red the moment you change the print statements — not because skip logic broke, but because the test's only way of checking behavior is scraping stdout text that the plan is about to change out from under it.

**Fix to implement:**
- [ ] Before changing any print statements, update `skip_test.go`'s assertions to match the new output strings (or, better — see B4 below — stop asserting on printed text at all).
- [ ] Treat this as step 1, not an afterthought — "does the test suite still pass" should gate this change, not follow it.

### B2. Non-deterministic task ordering will undermine the exact "premium" feel you're going for

**The problem, unrelated to the logging plan itself but exposed by it:** `dag.go` builds each tier by iterating `g.Nodes`, which is a Go map — iteration order is randomized per run. That means the new `[ Orbit ] Running format, lint...` header (and the order checkmarks print in) will vary from run to run: sometimes `format, lint`, sometimes `lint, format`. The current boring `Executing Tier 0 (2 tasks)...` line never showed task names, so this was invisible before. The new format puts task names front and center, which makes the flakiness visible for the first time — and inconsistent ordering is exactly the kind of thing that makes a CLI feel janky rather than premium, which undercuts the stated goal.

**Fix to implement (`pkg/graph/dag.go`):**
- [ ] Sort nodes alphabetically by name (or another stable rule) when building each tier in `TopologicalSort`, so both the "Running X, Y..." header and the per-task result lines print in a consistent order across runs.
- [ ] Do this before or alongside the logging change — implementing the new format on top of non-deterministic ordering means you'll ship the "premium" output and then immediately notice it looks different every time you run it.

### B3. Emoji/Unicode symbols on Windows need an explicit check, not an assumption

**The problem:** `✔`, `✖`, `🚀`, `✨` render fine in modern terminals (Windows Terminal, most Unix shells) but can come out as mojibake in legacy `cmd.exe` or older PowerShell hosts without UTF-8 code page configured — and your CI matrix already explicitly tests `windows-latest`, which is exactly the environment where this could silently look broken for a meaningful chunk of your Windows users.

**Questions to answer:**
- Have you actually run this in a plain Windows `cmd.exe` window (not Windows Terminal) to confirm the symbols render? If not, that's a five-minute check worth doing before committing to Unicode icons as the default.

**Fix to implement:**
- [ ] Test the proposed output in both Windows Terminal and legacy `cmd.exe`.
- [ ] If legacy rendering is broken, either set the UTF-8 code page programmatically on Windows startup, or fall back to ASCII (`[OK]`/`[FAIL]`/`[SKIP]`) when `runtime.GOOS == "windows"` and no UTF-8 support is detected.

### B4. Bigger architectural suggestion — decouple output formatting from execution logic now

**Why this matters beyond just this change:** your own plan asks "should we use spinners later (pterm/huh)?" as an open question. Right now, `runner.go` prints directly via `fmt.Printf` inline with execution logic, and your tests verify behavior by capturing and parsing stdout (`skip_test.go`'s `os.Pipe()` redirect). That means *every* future output change — spinners, `--quiet` mode, JSON output for CI, this logging revamp itself — has to touch `runner.go`'s core logic and risks breaking tests that were never meant to be about formatting.

**Suggested fix (bigger than this one plan, worth deciding now rather than after a third output revamp):**
- [ ] Introduce a small `Reporter` interface (`TaskStarted`, `TaskSkipped(reason)`, `TaskSucceeded`, `TaskFailed(output)`, `TierStarted(names)`) that `ExecuteTiers` calls into instead of printing directly.
- [ ] Ship one `TextReporter` implementation now with the format from this plan. This costs you maybe an extra hour today, and it's what makes B3's Windows fallback and a future spinner/JSON reporter each a new implementation of the interface rather than a rewrite of `runner.go`.
- [ ] Update tests to assert against a `TaskResult`/`Status` struct (which already exists) or a mock `Reporter`'s recorded calls, instead of parsing printed strings — this is what actually fixes the fragility that made B1 possible in the first place, not just this one instance of it.

---

## Suggested order
1. **A1, A2** — these are correctness bugs in already-shipped behavior; fix before layering new UI on top of a runner whose skip logic has known bugs.
2. **B2 (deterministic ordering)** — small, and doing it before the logging change means you're not shipping a "premium" format that's visibly inconsistent on day one.
3. **B4 (Reporter interface)** — worth the extra hour now; then B1's test updates become part of writing `TextReporter`, and the format change from the original plan becomes step 4, not step 1.
4. **B1 (update tests) / apply the format change** — via the `TextReporter` from B4.
5. **B3 (Windows rendering check)** — verify before calling this done, since your CI already tests Windows and a rendering regression there would be embarrassing to ship unnoticed.
6. **A3** — small polish, fits in whenever.
