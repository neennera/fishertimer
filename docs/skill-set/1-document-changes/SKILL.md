---
name: document-changes
description: Guidelines and automated protocol for AI agents to document every code, schema, and config change into /docs/CHANGELOG.md.
---

# Skill 1: Document Changes

## Objective
Ensure that every change made to the codebase—whether by human engineers or autonomous AI agents—is accurately logged in [`/docs/CHANGELOG.md`](../../CHANGELOG.md) adhering to the [Keep a Changelog](https://keepachangelog.com/) standard.

---

## 1. When to Document Changes
You MUST document changes whenever you:
1. Add a new service, package, endpoint, or feature.
2. Modify an existing API contract, domain entity, or database schema.
3. Fix a bug, regression, or environment issue.
4. Refactor internal code or service layer boundaries.
5. Add or update configurations, scripts, or build pipelines.

---

## 2. Changelog Structure & Sections

Under the `## [Unreleased]` heading in [`/docs/CHANGELOG.md`](../../CHANGELOG.md), group modifications into the standard categories:

- **`### Added`**: For new user-facing features, microservices, endpoints, or packages.
- **`### Changed`**: For changes in existing functionality, interfaces, or configs.
- **`### Deprecated`**: For once-stable features soon to be removed.
- **`### Removed`**: For deprecated features removed in this release.
- **`### Fixed`**: For bug fixes, compile errors, or environmental corrections.
- **`### Security`**: For vulnerability patches or authentication/authorization fixes.

---

## 3. Entry Format Checklist

Each changelog item should follow this concise format:
- Specify the affected scope/service in bold (e.g. **`[study-timer]`** or **`[web]`**).
- Describe *what* changed and *why* (avoid trivial statements like "edited file").
- Link relevant docs or PRs if applicable.

### Example Entry:
```markdown
## [Unreleased]

### Added
- **[study-session]**: Implemented `POST /api/sessions/join` handler with atomic capacity checking.
- **[shared-types]**: Added `JoinSessionRequest` and `JoinSessionResponse` DTO contracts.

### Fixed
- **[turbo.json]**: Whitelisted `%LocalAppData%` and `%GOCACHE%` in `globalPassThroughEnv` to fix Go compilation failures on Windows.
```

---

## 4. Execution Step for Agents
1. Before finishing your task, check your git status or diff using `git status --short`.
2. Open [`/docs/CHANGELOG.md`](../../CHANGELOG.md).
3. Insert bullet points under `## [Unreleased]` in the appropriate category.
4. Verify formatting matches existing entries.
