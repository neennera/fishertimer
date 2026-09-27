# Lessons Learned & Agent Mistake Log

> **Notice to AI Agents:** This living log records encountered errors, environmental quirks, architectural edge cases, and actionable solutions. Before diagnosing strange build failures or runtime bugs, review this file to prevent repeating past mistakes.
>
> **Agent Protocol:** Whenever an agent discovers a non-obvious bug or failure mode and solves it, document the incident here following the template in Section 2.

---

## 1. Incident & Knowledge Archive

### Incident 001: Windows PowerShell Script Execution Policy Block
- **Date:** 2026-09-11
- **Component:** Windows CLI / pnpm / npm
- **Symptom:** Running `pnpm`, `npm`, or `turbo` in PowerShell aborted with:
  `File C:\Users\...\pnpm.ps1 cannot be loaded because running scripts is disabled on this system.`
- **Root Cause:** Windows Client PowerShell execution policy defaults to `Restricted`, which blocks `.ps1` wrapper scripts generated in `AppData\Roaming\npm`.
- **Solution:** Execute package manager commands via `cmd /c "<command>"` or invoke `powershell -NoProfile -ExecutionPolicy Bypass -Command "<command>"`.
- **Action for Future Agents:** Never rely on raw PowerShell script invocation on Windows without checking or bypassing execution policies.

---

### Incident 002: Go Build Fails under Turborepo v2 Strict Environment Filtering
- **Date:** 2026-09-11
- **Component:** Turborepo v2 `build` task & Go compiler
- **Symptom:** Running `pnpm turbo build` on Go microservices failed with:
  `build cache is required, but could not be located: GOCACHE is not defined and %LocalAppData% is not defined`
- **Root Cause:** Turborepo v2 operates in `strict` environment mode by default and strips all environment variables not explicitly configured, hiding `%LocalAppData%` and `%GOCACHE%` from the Go toolchain.
- **Solution:** Add `globalPassThroughEnv` in `turbo.json`:
  ```json
  "globalPassThroughEnv": [
    "LocalAppData",
    "LOCALAPPDATA",
    "GOPATH",
    "GOROOT",
    "GOCACHE",
    "PATH",
    "USERPROFILE",
    "SystemRoot",
    "HOME"
  ]
  ```
- **Action for Future Agents:** Whenever adding non-Node compilers (e.g. Go, Rust, Python) to a Turborepo pipeline, ensure compiler cache and OS paths are explicitly declared in `globalPassThroughEnv`.

---

### Incident 003: Go Workspaces Resolution with Subdirectory Relative Paths
- **Date:** 2026-09-11
- **Component:** Go 1.22 multi-module workspace (`go.work`)
- **Symptom:** Running `go build ./services/...` from the monorepo root failed with:
  `directory prefix services does not contain modules listed in go.work`
- **Root Cause:** In Go workspaces, modules listed in `go.work` are independent modules. Wildcard path matching from root does not automatically cross module boundaries without using the explicit module path or running inside each module.
- **Solution:** Use Turborepo task filtering (`pnpm turbo build --filter=@fishertimer/*-service`) or run commands per module directory (`cd services/<svc> && go build`).
- **Action for Future Agents:** Rely on Turborepo as the orchestrator to navigate into workspace packages rather than guessing multi-module glob patterns.

---

### Incident 004: gRPC Requires a Newer Go Than the Workspace Declared
- **Date:** 2026-09-27
- **Component:** `go.work`, `proto` module, `google.golang.org/grpc`
- **Symptom:** The workspace declared `go 1.22`, but the current `google.golang.org/grpc` cannot be used at that version; `go mod init` / `go mod tidy` also silently set the new module's `go` directive to the local toolchain (`1.27.1`), which would force every teammate onto that exact version.
- **Root Cause:** `google.golang.org/grpc` v1.84 declares `go 1.25.0` in its own `go.mod`. A dependency's minimum Go version propagates to every module that imports it, and `go.work` must be at least as new as each member module.
- **Solution:** Raised `go.work` and the modules that import gRPC (`proto`, `study-timer`, `api-gateway`) to `go 1.25.0`. Contributors need Go 1.25 or newer installed.
- **Action for Future Agents:** Before adding a Go dependency, check the `go` directive in that dependency's `go.mod`; if it is newer than `go.work`, raise the workspace deliberately (and tell the team) rather than letting `go mod tidy` do it silently.

---

### Incident 005: Shared Proto Module Must Also Build Outside the Workspace
- **Date:** 2026-09-27
- **Component:** `proto` module consumed by `study-timer` and `api-gateway`
- **Symptom:** Importing `github.com/neennera/fishertimer/proto/...` works under `go.work`, but `GOWORK=off go build` (as a deploy target such as Render builds a single service) cannot find the module, because it is not published anywhere.
- **Root Cause:** `go.work` only resolves local modules during workspace builds; a standalone module build resolves imports from its own `go.mod` and the module proxy.
- **Solution:** Each consumer's `go.mod` requires `github.com/neennera/fishertimer/proto v0.0.0` with `replace github.com/neennera/fishertimer/proto => ../../proto`, then `GOWORK=off go mod tidy`.
- **Action for Future Agents:** When a service imports a local workspace module, add the `require` + `replace` pair and verify with `GOWORK=off go build ./...`.

---

## 2. Template for Recording New Lessons Learned

When documenting a new learning, append to Section 1 using this markdown structure:

```markdown
### Incident <Number>: <Short Descriptive Title>
- **Date:** YYYY-MM-DD
- **Component:** <Affected service, package, or tool>
- **Symptom:** <Exact error message or undesirable behavior>
- **Root Cause:** <Technical reason why the failure occurred>
- **Solution:** <Code or configuration change that fixed the problem>
- **Action for Future Agents:** <Specific rule or heuristic to avoid this in future turns>
```
