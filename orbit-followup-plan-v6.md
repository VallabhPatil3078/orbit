# Orbit — Follow-Up Plan v6

The three items deferred from v5: full timeout/signal design, versioned releases, and config schema versioning. Each section below tries to surface the edge cases *now* rather than after they're discovered in the wild — that's the whole point of not wanting a v7 for these. Where a decision is genuinely open rather than a clear "right answer," it's called out as a question rather than a prescribed fix.

---

## 1. Full task-timeout & signal design

v5 shipped a working global timeout with a real test. This section is everything that was explicitly deferred: per-task config, graceful-vs-forceful kill, and cross-OS process-tree cleanup — plus the edge cases each one raises.

### 1.1 Per-task configurable timeout

**Fix to implement:**
- [ ] Add `timeout: "30s"` (or `timeout_seconds: 30` — pick one, don't support both) to `TaskConfig` and `graph.Node`, parsed as a `time.Duration`.
- [ ] Fall back to the existing global default (`10 * time.Minute`, or make that configurable at the top level of `orbit.yaml` too) when a task doesn't specify its own.

**Edge cases to cover:**
- [ ] **Invalid values.** `timeout: -5s`, `timeout: 0s`, or an unparseable string (`timeout: "soon"`) — these must fail in `orbit validate`, not silently become "no timeout" or "instant timeout." Add a test for each.
- [ ] **Absurdly large values.** Not a hard error, but consider a sanity warning (not a failure) if a timeout exceeds some threshold like 1 hour — more likely a typo (`30` meant as minutes, parsed as `30ns`) than an intentional value. Optional, but cheap to add and catches a real class of config typo.
- [ ] **Zero tasks with any timeout defined at all.** Confirm the global default still applies when *no* task in the config sets one — i.e. this feature is purely additive, not a config requirement.

### 1.2 Graceful-then-forceful kill

**The problem today:** `context` cancellation via `exec.CommandContext` sends an immediate kill (`SIGKILL` on Unix) with no chance for the process to clean up (close file handles, flush output, remove a lock file). A build tool killing `mvn` mid-write is more likely to leave a corrupted half-written artifact than a tool that gives it a moment to exit cleanly.

**Questions to answer:**
- What's an acceptable grace period? A common default is 5–10 seconds between "please stop" (`SIGTERM`) and "stop now" (`SIGKILL`) — does that fit your use case, or do fast pre-commit tasks want a shorter window?

**Fix to implement:**
- [ ] On Unix: send `SIGTERM` first, start a short grace-period timer, then `SIGKILL` if the process hasn't exited by the time it elapses.
- [ ] On Windows: there's no real equivalent to `SIGTERM` for arbitrary processes — Windows either lets a process handle `WM_CLOSE` (GUI apps) or you terminate it outright. For console processes like build tools, a "graceful" stop usually isn't meaningfully available; document this platform difference rather than pretending parity exists, and just do a direct terminate on Windows.

**Edge cases to cover:**
- [ ] **The process ignores `SIGTERM` entirely** (some processes do, intentionally or via a bug) — confirm the grace-period timer reliably escalates to `SIGKILL` and doesn't hang waiting forever.
- [ ] **The process exits normally in the small window between deadline and `SIGTERM` being sent** — make sure this isn't misreported as a timeout when it was actually about to finish on its own. (This is the same "near-simultaneous completion vs. cancellation" race from v5 — worth a dedicated test now that there's a multi-step kill sequence instead of one signal.)

### 1.3 Cross-OS process-tree cleanup (the part most likely to bite you if skipped)

**The problem:** `exec.CommandContext` only kills the *direct* child — `sh` on Unix, `cmd.exe` on Windows. If a task's command spawns its own children (`npm test` forking a test runner, a shell script backgrounding a process with `&`), those grandchildren are not guaranteed to die when the parent is killed. This is exactly the gap flagged when the v5 timeout fix shipped — v5 solved "the immediate process won't hang forever," not "nothing is ever orphaned."

**Fix to implement:**
- [ ] **Unix:** set `SysProcAttr{Setpgid: true}` on the command so it starts its own process group, then on timeout/cancel send the kill signal to the *negative* PID (`-pid`), which targets the whole group instead of just the one process.
- [ ] **Windows:** direct child killing doesn't propagate to grandchildren the way Unix process groups do. Use a Job Object (`CreateJobObject` + `AssignProcessToJobObject` + `TerminateJobObject` via `golang.org/x/sys/windows`, or a small wrapper package) so the whole tree dies together. This is meaningfully more work than the Unix side — budget for it accordingly rather than assuming it's symmetric.

**Edge cases to cover:**
- [ ] **A grandchild that detaches/daemonizes itself** (rare, but real — e.g. a process that explicitly forks and exits its parent to survive independently) — decide and document that Orbit's guarantee is "processes in the tree at kill-time are terminated," not "anything the task ever spawned, forever." Trying to guarantee the latter is a much bigger problem (full process supervision) and out of scope.
- [ ] **Integration test, not unit test:** write a test task whose command starts a background process and exits immediately (e.g. `sh -c "sleep 100 & exit 0"` on Unix, an equivalent on Windows), let Orbit's tier think the task succeeded fast, then trigger a pipeline-level cancel/timeout and assert the backgrounded `sleep` is actually gone afterward (check the PID no longer exists). This is the one test in this whole plan that actually proves the feature works — the earlier `TestExecuteTiers_Timeout` from v5 only proved the *direct* process dies, not the tree.

### 1.4 Sibling-task behavior when one task times out (open design question — decide before implementing)

**The problem:** today, when one task in a tier times out, `wg.Wait()` still waits for every other goroutine in that tier to finish normally — a sibling task with no timeout of its own keeps running to completion even though its tier-mate already failed.

**Questions to answer:**
- Should a timeout in one task cancel its siblings in the same tier immediately (fail-fast — stop wasting time on work that's going to be discarded anyway), or let them finish (current behavior — you get full information about what else would have passed/failed)?
- If you fail-fast, siblings that get cancelled as a side effect need their own distinct status — they didn't time out and they didn't fail on their own merits, they were cancelled because a sibling did. Reusing `StatusFailed` for this would be misleading in the output.

**Fix to implement (once the above is decided):**
- [ ] If fail-fast: derive a per-tier `context.WithCancel` from the pipeline context, and call its cancel function as soon as any task in the tier hits `StatusFailed` or times out — the remaining goroutines' `exec.CommandContext` calls pick this up and exit early.
- [ ] Add a `StatusCancelled` (distinct from `StatusFailed`/`StatusSkipped`) if you go this route, and a corresponding `Reporter.TaskCancelled` method.
- [ ] Test both the "one task times out, sibling with a long-but-under-timeout command" case and confirm the chosen behavior (either the sibling completes, or it's cleanly cancelled — whichever was decided above).

---

## 2. Versioned releases

**The problem today:** "installing Orbit" means cloning the repo and running `go build` yourself. No git tags, no changelog, no published binaries, and — worth checking — there's currently no `orbit --version` flag, so even someone who did build it has no way to confirm what they're running.

### 2.1 Version embedding in the binary

**Fix to implement:**
- [ ] Add a `version` package-level variable in `main.go`, injected at build time via `-ldflags "-X main.version=$(git describe --tags --always)"`.
- [ ] Add an `orbit --version` / `orbit version` command that prints it.
- [ ] Default to something explicit like `"dev"` when built without the ldflag (a plain `go build` from source) so it's never ambiguous whether a binary is a tagged release or a local build.

### 2.2 Tags, changelog, and a release workflow

**Fix to implement:**
- [ ] Adopt semver tags (`v0.1.0`, etc.) — given the project is pre-1.0 and still changing behavior (like the cascade-skip fix in v4), stay in the `0.x` range until the config/skip-logic semantics are ones you're confident won't change again.
- [ ] Add `CHANGELOG.md` in Keep a Changelog format, updated per tag — this is also just useful to *you*, since this whole review process has already produced several behavior changes (cascade AND/OR fix, exit codes, timeout semantics) that a changelog would have made trivial to summarize instead of re-deriving from commit messages.
- [ ] Add a second GitHub Actions workflow (separate from your existing CI workflow), triggered on tag push (`on: push: tags: ["v*"]`), that cross-compiles for the common `GOOS`/`GOARCH` combinations (`linux/amd64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`) and uploads the binaries to a GitHub Release. GoReleaser is the standard tool for this if you want to avoid hand-rolling the matrix.

**Edge cases to cover:**
- [ ] **A tag pushed to a commit that hasn't passed the normal CI workflow.** Either make the release workflow depend on CI success (via `workflow_run` or just re-running the test matrix as a release-workflow step before building), or document that tags should only ever be pushed after CI is green on `main` — don't let a release ship untested.
- [ ] **The module path.** `go.mod` currently declares `module orbit` — a short, non-VCS-rooted name. For `go install github.com/<you>/orbit@latest` to work for anyone else, the module path needs to match the actual repo path (`module github.com/<you>/orbit`). This is a breaking rename for your own imports (`"orbit/pkg/graph"` → `"github.com/<you>/orbit/pkg/graph"` everywhere) — do it once, deliberately, before your first real tag, not after, since changing it later invalidates any `go install` pins people have already made.

---

## 3. Config schema versioning

**The problem:** `orbit.yaml` has already changed shape multiple times across this review (`working_dir`, `trigger_paths`, `ignore_paths`, and a behavior change in what "no filters" means for cascade skip) with no field marking which version of the schema a given file was written against, and no compatibility policy stated anywhere.

### 3.1 Add and enforce a version field

**Fix to implement:**
- [ ] Add a top-level `version: 1` field to the `orbit.yaml` schema (config struct + the shipped example file).
- [ ] In `ParseConfig`, check it explicitly rather than ignoring unknown fields.

**Edge cases to cover:**
- [ ] **Every `orbit.yaml` written before this change has no `version` field at all.** This must not become a hard error for existing users — treat a missing `version` as `version: 1` (today's schema) implicitly, so this change itself doesn't break anyone who already adopted the tool during this review process.
- [ ] **A config declares a version newer than the binary understands** (e.g. someone shares a `v2` config, but the recipient has an older Orbit binary that only knows `v1`). This should produce a clear, specific error — `"this config requires Orbit schema v2 or later; you're running a v1-compatible build"` — not a confusing field-parsing failure that looks like a typo in the YAML.
- [ ] **A config declares a version the binary is newer than and no longer fully supports** (post-1.0, if you ever drop support for a very old schema version). Decide now, even if it doesn't apply yet: do you want to support old schema versions indefinitely, or reserve the right to drop them on a major version bump? Write this down as a stated policy (even one sentence in the README) rather than leaving it implicit — that's the actual goal of versioning the schema in the first place.

### 3.2 Document the compatibility contract

**Fix to implement:**
- [ ] Add a short section to the README stating the policy in concrete terms, e.g.: "Within the same major schema version, fields are only ever added, never removed or repurposed. A field's meaning, once shipped, doesn't change without a version bump." This is the promise that makes the version number actually mean something to a user, rather than being a number nobody checks.

---

## What's intentionally still not "done" after this, and why that's okay
Two things in this plan aren't one-time code fixes, even once implemented — they're ongoing practices, not bugs to close out:
- **Cutting releases** is something you'll keep doing as the project evolves, not a task with a finish line.
- **The changelog and compatibility policy** only have value if kept up going forward, the same way tests only have value if kept passing.

Everything else above — per-task timeouts, graceful kill, process-tree cleanup, the sibling-cancellation decision, version embedding, the schema version field — is scoped to be genuinely finishable in this pass, with the edge cases named up front rather than left to be discovered. If something here does spill into a v7, it should only be 1.3 (process-tree cleanup) or 1.4 (the sibling-cancellation decision) — those are the two with real design judgment calls and platform-specific work, not just missing code.
