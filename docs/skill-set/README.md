# Fisher Timer AI Agent Skill Suite

Welcome to the Fisher Timer Agent Skill Suite. Because this codebase is primarily developed, maintained, and evolved by AI Agents, these standardized skills ensure consistency, documentation integrity, and error prevention across all agent turns.

---

## Available Skills

| ID | Skill Directory | Name | Purpose |
| :--- | :--- | :--- | :--- |
| **0** | [`0-explain-repo`](0-explain-repo/SKILL.md) | **Explain This Repo** | Fast onboarding for agents: architecture, tech stack, ADRs, port map, and code layout. |
| **1** | [`1-document-changes`](1-document-changes/SKILL.md) | **Document Changes** | Protocol for appending structured, semantic updates to [`/docs/CHANGELOG.md`](../CHANGELOG.md). |
| **2** | [`2-update-architecture`](2-update-architecture/SKILL.md) | **Update Architecture** | Step-by-step instructions to keep [`/docs/ARCHITECTURE.md`](../ARCHITECTURE.md) and diagrams synchronized. |
| **3** | [`3-way-of-work`](3-way-of-work/SKILL.md) | **Way of Work & Lessons** | Testing standards, build verification, and logging bugs/learnings in [`/docs/LESSONS_LEARNED.md`](../LESSONS_LEARNED.md). |
| **4** | [`4-database-and-repository`](4-database-and-repository/SKILL.md) | **Database & Repository** | Repository pattern in Clean Architecture, schema locations, sample models, and DB adapter guidelines. |
| **5** | [`5-migration-and-seeding`](5-migration-and-seeding/SKILL.md) | **Migration & Seeding** | Docker Compose database lifecycle (`pnpm db:up/down/reset`), migration workflows, and seed data. |

---

## Agent Execution Rule

Whenever an agent begins a task:
1. Review [`0-explain-repo`](0-explain-repo/SKILL.md) to understand the relevant services.
2. If working with data persistence, review [`4-database-and-repository`](4-database-and-repository/SKILL.md) and [`5-migration-and-seeding`](5-migration-and-seeding/SKILL.md).
3. Follow [`3-way-of-work`](3-way-of-work/SKILL.md) for clean layer separation and verification tests.
4. If new APIs, databases, or services are added/altered, execute [`2-update-architecture`](2-update-architecture/SKILL.md).
5. Before finishing the turn, execute [`1-document-changes`](1-document-changes/SKILL.md) to document changes in [`/docs/CHANGELOG.md`](../CHANGELOG.md).
