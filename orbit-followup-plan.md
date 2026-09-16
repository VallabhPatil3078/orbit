# Orbit — Follow-Up Plan (from code review)

Five gaps were found in the review. Each is framed as a question to answer/decide, with the concrete fix underneath. Tackle in this order — top to bottom is roughly "highest real-world risk first."

---

## 1. `orbit init` silently overwrites existing files

**Questions to answer:**
- If `orbit.yaml` already exists in this directory, what should `orbit init` do — abort with an error, skip silently, or prompt to confirm?
- Same question for `.git/hooks/pre-commit` — if a user already has a pre-commit hook (from Husky, pre-commit.com, or their own script), should Orbit refuse to touch it, back it up, or append itself to it?
- Do you want a `--force` flag to opt into overwriting, so the safe path is the default?

**Fix to implement (`cmd/init.go`):**
- [ ] Before `os.WriteFile("orbit.yaml", ...)`, check `os.Stat("orbit.yaml")` — if it exists, print `[!] orbit.yaml already exists, skipping (use --force to overwrite)` and return, unless `--force` was passed.
- [ ] Before writing `hookPath`, check if it exists. If it does and doesn't already contain `orbit run`, either refuse and tell the user to add `orbit run` manually, or rename the existing hook to `pre-commit.orbit-backup` first.
- [ ] Add a `--force` bool flag to `initCmd` via `initCmd.Flags().BoolVar(...)`.

---

## 2. No test for a missing/typo'd dependency

**Questions to answer:**
- What's the exact user-facing error message you want when `depends_on` references a task name that doesn't exist in the config? Is `"dependency not found: " + dep` (current) good enough, or should it also name *which task* has the bad reference (e.g. `task "build" depends on unknown task "lnit"`)?

**Fix to implement (`pkg/graph/dag_test.go`):**
- [ ] Add `TestBuildEdges_MissingDependency`: create a DAG where one node's `DependsOn` points to a name never added via `AddNode`, call `BuildEdges()`, assert the error is non-nil and mentions the missing name.
- [ ] Optionally improve `dag.go`'s error to include the *dependent* task's name, not just the missing dependency's name — this is the more useful message when debugging a real `orbit.yaml`.

---

## 3. No diamond-dependency test

**Questions to answer:**
- Do you want this as one test or two — one diamond (A → B,C → D) and one wider fan-out/fan-in (A → B,C,D → E)?

**Fix to implement (`pkg/graph/dag_test.go`):**
- [ ] Add `TestTopologicalSort_Diamond`: nodes `A` (no deps), `B` and `C` (both depend on `A`), `D` (depends on both `B` and `C`). Assert 3 tiers: `[A]`, `[B, C]`, `[D]`. This is the case most likely to expose an in-degree counting bug that a simple chain wouldn't catch.

---

## 4. Thin config validation

**Questions to answer:**
- Should an empty `command: ""` be a hard config error (fail before running anything), or is failing at `cmd.Run()` time acceptable?
- Should task names be validated (no spaces, no special chars) since they're used as map keys and in adjacency lists?
- Do you want a `orbit validate` subcommand that just parses + builds the DAG without running anything — useful in CI before a real commit?

**Fix to implement (`pkg/config/parser.go`):**
- [ ] After `yaml.Unmarshal`, loop over `cfg.Tasks` and return an error for any task with `Command == ""`.
- [ ] Add a test in a new `parser_test.go`: malformed YAML → error; empty command → error; valid config → no error.
- [ ] (Optional, nice-to-have) `orbit validate` command in `cmd/` that runs steps 1–3 of `run.go` (parse, build edges, topo sort) and reports success/failure without calling `runner.ExecuteTiers`.

---

## 5. No test that tasks in a tier actually run concurrently

**Questions to answer:**
- What's an acceptable way to assert concurrency in a test — timing-based (two `sleep 1` tasks finish in ~1s total, not ~2s) or instrumentation-based (each task writes a timestamp, assert overlap)?

**Fix to implement (`pkg/runner/runner_test.go`):**
- [ ] Add `TestExecuteTiers_RunsConcurrently`: two nodes with `Command: "sleep 1"` in the same tier (use `sh -c "sleep 1"`/cross-platform equivalent), time the call to `ExecuteTiers`, assert total duration is well under 2 seconds (e.g. `< 1.5s`) to prove they ran in parallel, not sequentially.

---

## Stretch (not blocking, but worth a line in the README if skipped)
- [ ] What happens if two tasks in the same tier fail? Right now only the first error read off the channel is returned/printed — the second failing task's output is silently dropped. Decide: collect and print all failures in a tier, or is first-failure-wins an intentional simplification worth documenting?
