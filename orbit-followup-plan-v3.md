# Orbit — Follow-Up Plan v3: Context-Aware Orchestration (`trigger_paths` / `ignore_paths`)

Before writing the glob-matching engine, nail down these design decisions — each one below is a place the original brainstorm left ambiguous, and ambiguity here means silently-wrong skip decisions later, which are much harder to debug than a crash.

---

## 1. Path relativity — define `working_dir` explicitly

**The problem:** `git diff` outputs repo-root-relative paths. Tasks currently bury `cd backend && ...` inside the command string, so it's undefined whether `trigger_paths` globs are meant to be repo-root-relative or relative to wherever the command `cd`s into.

**Questions to answer:**
- Do you want `trigger_paths`/`ignore_paths` to always be repo-root-relative (simplest, consistent with git's own output), or relative to a per-task `working_dir`?

**Fix to implement:**
- [ ] Add an explicit `working_dir` field to `TaskConfig` and `graph.Node`, separate from the `command` string (e.g. `working_dir: "backend"` instead of `command: "cd backend && mvn ..."`).
- [ ] Decide and document: `trigger_paths`/`ignore_paths` are always evaluated against repo-root-relative git diff output, regardless of `working_dir`. `working_dir` only affects where the command executes, not how paths are matched.
- [ ] Add a test with a task whose `working_dir` differs from repo root, confirming trigger matching still works against root-relative diff paths.

---

## 2. Handle the no-HEAD (first commit) case

**The problem:** `git diff --cached --name-only` compares the index to `HEAD`. On a brand-new repo with zero commits, `HEAD` doesn't exist, and this errors out (`ambiguous argument 'HEAD'`). Anyone running `orbit init` as their first step, then committing, hits this immediately.

**Fix to implement:**
- [ ] Before running the diff, check whether `HEAD` exists (e.g. `git rev-parse --verify HEAD` — non-zero exit means no commits yet).
- [ ] If no `HEAD`, diff against git's empty-tree hash (`4b825dc642cb6eb9a060e54bf8d69288fbee4904`) instead, or treat "no HEAD" as "run everything unconditionally" — pick whichever you find less surprising and document the choice.
- [ ] Add a test that initializes a fresh temp git repo with zero commits, stages a file, and confirms Orbit doesn't crash.

---

## 3. Cascade skip vs independent evaluation — resolve per-task, not globally

**The decision:** don't pick Option A or B as a single global mode. Instead:
- If a child task defines its **own** `trigger_paths`/`ignore_paths`, evaluate it independently of whether its parent ran or was skipped (Option B logic) — the child is explicitly declaring its own trigger criteria.
- If a child task defines **no** path filters of its own, it implicitly depends entirely on its parent, so cascade-skip it when the parent is skipped (Option A logic).

**Fix to implement (`pkg/runner/runner.go`):**
- [ ] Add a `Skipped` result state alongside `Success`/`Failed` (don't overload `error == nil` to mean both "succeeded" and "skipped" — the tier summary and dependent-task logic both need to distinguish these).
- [ ] Before executing a tier, for each node: if the node has its own trigger/ignore paths, evaluate against the diff directly. If it has none, inherit its parent's skip status (skip if any parent was skipped).
- [ ] Add tests for both branches: (a) child with its own `trigger_paths` runs even though its parent was skipped, (b) child with no path filters is cascade-skipped when its parent is skipped.

---

## 4. Define precedence when both `trigger_paths` and `ignore_paths` are set on the same task

**Questions to answer:**
- If both are defined, should `ignore_paths` be applied first (excluding matches), with `trigger_paths` then required to match at least one remaining changed file? That's the more intuitive reading, but confirm it's what you want before locking in behavior.

**Fix to implement:**
- [ ] Document the precedence rule directly in the YAML schema comments/README, not just in code.
- [ ] Add a test with both fields set on one task, covering: a file that matches `trigger_paths` but is also excluded by `ignore_paths` (should skip), and a file that matches `trigger_paths` and isn't excluded (should run).

---

## 5. Add escape hatches for CI and non-hook usage

**The problem:** the staged-files diff only makes sense in the pre-commit-hook context. A CI checkout typically has nothing staged, so without a fallback, this feature does nothing useful outside of local commits.

**Fix to implement (`cmd/run.go`):**
- [ ] Add a `--all` flag that bypasses skip logic entirely and forces every task to run — useful for CI and for debugging "why did this skip" confusion.
- [ ] Add a `--base <ref>` flag that diffs against a given ref (e.g. `origin/main`) instead of the git index — this is what makes the feature usable in CI, where you'd diff the PR branch against its merge-base rather than relying on staged files.
- [ ] Add a test confirming `--all` produces the same tier execution as today's behavior (no skips at all), for backward compatibility.

---

## 6. Integration tests (this is the real testing lift, budget time for it)

**Why this is different from your existing test suite:** everything you've tested so far (topo sort, cycle detection, task failure) is pure in-memory logic. This feature depends on real git state, which means unit tests with mocked structs won't catch the bugs that matter (path relativity, no-HEAD, precedence).

**Fix to implement (new `pkg/runner/gitdiff_integration_test.go` or similar):**
- [ ] Spin up a real temp directory, run `git init`, create and stage specific files, and assert which tasks Orbit decides to run vs skip — for at least: a single trigger match, a single ignore match, the combined trigger+ignore case from item 4, the no-HEAD case from item 2, and the cascade-vs-independent case from item 3.
- [ ] These will be slower than your current unit tests (real subprocess calls to `git`). Consider a build tag or a separate `go test -tags=integration` target so they don't slow down the fast feedback loop of your normal test run.

---

## Suggested order
1. **Item 1 (`working_dir`)** and **item 2 (no-HEAD)** first — both are small, foundational, and everything else builds on correct path handling.
2. **Item 4 (precedence)** — cheap to decide and test now, expensive to change later once tasks in the wild rely on one behavior.
3. **Item 3 (skip state + cascade logic)** — the core mechanic; do this once the groundwork above is solid.
4. **Item 5 (`--all` / `--base` flags)** — needed before this is usable anywhere but a local pre-commit hook.
5. **Item 6 (integration tests)** — write alongside item 3, not after; the cascade logic is exactly the part that's hardest to get right without real git fixtures to test against.
