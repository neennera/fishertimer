# Fisher Timer - C4 diagrams (PlantUML + C4-PlantUML)

| File | C4 level | Shows |
| :--- | :--- | :--- |
| `c4_level2_container.puml` | Level 2 - Container | All containers: Web App, API Gateway, 6 microservices, 5 databases + Redis, RabbitMQ, optional Consul, and every sync (HTTP / gRPC) and async (AMQP) link |
| `c4_level3_api_gateway.puml` | Level 3 - Component | Identity middleware, route table, reverse proxies, REST to gRPC handlers |
| `c4_level3_account.puml` | Level 3 - Component | OAuth, JWT, Postgres repo, timer / reward clients, profile.updated publisher |
| `c4_level3_study_session.puml` | Level 3 - Component | gRPC + HTTP handlers, sweeper, transactional outbox relay, profile consumer |
| `c4_level3_study_timer.puml` | Level 3 - Component | State machine, sweeper, idempotent RabbitMQ consumer, reward / session clients |
| `c4_level3_reward.puml` | Level 3 - Component | Award use case, rarity calculation, Mongo repo with in-memory fallback |
| `c4_level3_leaderboard.puml` | Level 3 - Component | Cache-validation use case, Redis repo, reward client, profile consumer |
| `c4_level3_admin.puml` | Level 3 - Component | Admin-only auth, moderation use case, session read / write clients, audit log |

The diagrams use the C4-PlantUML standard library (`!include <C4/C4_Container>`), so they need no internet access.

## Render

**Option 1 - VS Code:** install the "PlantUML" extension (needs Java + Graphviz), open a `.puml`, press `Alt+D`.

**Option 2 - command line:**

```bash
# needs Java and Graphviz (dot)
java -jar plantuml.jar -tpng -DPLANTUML_LIMIT_SIZE=16384 present_diagram/*.puml
# SVG (sharper for slides):
java -jar plantuml.jar -tsvg present_diagram/*.puml
```

**Option 3 - online:** paste a file into https://www.plantuml.com/plantuml/uml (the server is public, so avoid this if the code is sensitive).

## Notes

- Drawn from the code (`main.go` wiring, `pkg/events`, compose files), not only from `docs/ARCHITECTURE.md`.
- Differences from `docs/ARCHITECTURE.md`: Account calls Study Timer over HTTP (not gRPC), and Account / Study Session / Leaderboard also use RabbitMQ for `profile.updated` events (exchange `fisher.account`).
- Not drawn: Adminer (dev-only DB UI) and the Kubernetes Traefik ingress (deployment concern, a Level 2 deployment diagram would cover it).
