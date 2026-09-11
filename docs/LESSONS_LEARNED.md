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
