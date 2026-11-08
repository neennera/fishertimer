# Fisher Timer live demo (Docker)

This runbook is for the `demo-only` branch. The **Finish now (demo)** action is
available only when `DEMO_MODE=true` in Study Timer and
`NEXT_PUBLIC_DEMO_MODE=true` in the web build. The normal 15-minute reward rule
and production defaults are unchanged.

## 1. Get the demo code

The branch must be available in the Git remote before another Mac can fetch it.
If `origin/demo-only` exists, use these commands in an existing clone:

```bash
git status --short                 # save any local changes before switching
git fetch origin demo-only
git switch --track origin/demo-only
```

If `demo-only` already exists locally, use `git switch demo-only` instead. A
fresh clone or a checkout of `main` alone does **not** contain the demo action.
If `origin/demo-only` is absent, the branch owner must share the branch before
these checkout commands can work on another Mac. Do not push from this runbook.

## 2. Check the Mac

- Install and start Docker Desktop with Compose v2.24 or newer. Allow about
  4 GB of memory for Docker and enough free disk space for a full first build.
  Rehearse the entire stack before presentation day.
- Keep internet access for the first image build and Google sign-in.
- Stop other local Fisher Timer stacks. For the debug ports used below, make
  sure ports 3000, 8000, 8085, 8087, 8088, 5432–5435, 27017, 6379, 5672,
  and 15672 are free. If a host MongoDB server uses 27017, stop it. MongoDB
  Compass alone does not occupy that port.
- Node, pnpm, Go, DBeaver, and a host database installation are not needed
  when using the Docker Compose command below.

If an older Fisher Timer stack is running and pnpm is available, stop it first:

```bash
pnpm db:down
pnpm docker:down
pnpm k8s:down
```

Skip commands for stacks that have never been started on this Mac.

```bash
docker compose version
docker info --format 'Docker memory: {{.MemTotal}} bytes'
df -h .
```

## 3. Configure the demo

From the repo root, create `.env.docker` if it does not exist:

```bash
test -e .env.docker || cp .env.docker.example .env.docker
```

Edit that git-ignored file. Replace the example `GOOGLE_CLIENT_ID` and
`GOOGLE_CLIENT_SECRET` with real values, and add:

```dotenv
DEMO_MODE=true
NEXT_PUBLIC_DEMO_MODE=true
NEXT_PUBLIC_USE_MOCKS=false
TIMER_SWEEP_INTERVAL_SECONDS=2
```

In Google Cloud, the OAuth client's authorized redirect URI must include
`http://localhost:3000/api/auth/google/callback` exactly. Add both demo Google
accounts as test users if the app is in testing. Keep credentials out of Git
and messages. The web demo flag is baked in at build time, so rebuild after
changing it.

## 4. Start and check the stack

```bash
docker compose --env-file .env.docker -f docker-compose.yml -f docker-compose.dev.yml --profile app up -d --build
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile app ps
curl -fsS -o /dev/null -w 'web: %{http_code}\n' http://localhost:3000/
curl -fsS http://localhost:8000/health
curl -fsS http://localhost:8085/health
```

The first build can take several minutes. Wait for all services with health
checks to show `healthy` in `ps`; Adminer has no health check. With pnpm
installed, `pnpm docker:dev` and `pnpm docker:ps` run the same setup.

| URL | Purpose |
| --- | --- |
| http://localhost:3000 | App and Google sign-in |
| http://localhost:3000/admin | Live rooms (admin account) |
| http://localhost:3000/admin/logs | Moderation audit log |
| http://localhost:8000/health | API Gateway health |
| http://localhost:8085/health | Reward health |
| http://localhost:8088 | Adminer for PostgreSQL |

## 5. Rehearse the reward flow

1. Sign in as the room participant, finish first-time account setup if asked,
   and create a room.
2. Start a **15-minute work** timer. The button appears only while your own
   work timer is running. Press **Finish now (demo)**. This makes the cycle
   due, then Study Timer uses its normal completion, participant lookup,
   reward award, and reward-sent flow. It does not call Reward directly from
   the browser.
3. Check the catch reveal and the account fish tank.
4. For a two-person demo, sign in as the second participant in a **different
   browser profile or private window**, join the room, then start and finish
   another 15-minute work cycle. Tabs in the same browser profile share one
   `localhost` session cookie and will replace each other's login.
5. Confirm the timer log records `participant_count=2` for the second cycle:

   ```bash
   docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile app logs study-timer | grep 'demo cycle'
   ```

The databases are fresh on a new Mac. They do not inherit accounts, rooms, or
rewards from another computer.

## 6. Inspect data without DBeaver

Adminer at http://localhost:8088 connects to PostgreSQL inside the Compose
network. Choose **PostgreSQL**, user `postgres`, password `postgrespassword`:

| Server | Database | Inspect |
| --- | --- | --- |
| `timer-db` | `timer_db` | `cycles`: `COMPLETED`, `reward_status=SENT` |
| `account-db` | `account_db` | `users`: account roles |
| `session-db` | `session_db` | rooms and participants |
| `admin-db` | `admin_db` | `admin_logs`: moderation actions |

For MongoDB rewards, connect MongoDB Compass (if installed) to
`mongodb://mongoadmin:mongopassword@localhost:27017/reward_db?authSource=admin`
and inspect `user_rewards`. Or use the shell already in the Mongo container:

```bash
docker exec -it fishertimer-reward-db mongosh -u mongoadmin -p mongopassword --authenticationDatabase admin reward_db
```

```javascript
db.user_rewards.find().sort({ awarded_at: -1 }).limit(10).pretty()
```

The app also exposes `GET http://localhost:8000/api/reward/rewards?user_id=USER_UUID`.
Use the signed-in participant's `user_id` from `account_db.users`.

## 7. Use the admin pages

On a fresh database, Google sign-up creates a `CUSTOMER`. After completing
sign-up for the intended admin Google account, promote **only that account**
in this local demo database (replace the email below):

```bash
docker exec fishertimer-account-db psql -U postgres -d account_db -c "UPDATE users SET role='ADMIN' WHERE lower(email)=lower('admin@example.com') RETURNING display_name, role;"
```

Check that the command returned one row. Sign out and back in so the new JWT
contains `ADMIN`. Keep the admin account in a different browser profile from
the room participant. The audit log records admin removals and room closures;
timer rewards do not create audit rows.

## If a step fails

| Symptom | Check |
| --- | --- |
| `invalid_client` at Google | `.env.docker` still has example credentials, or the account container was not recreated after editing it. Rerun the `up -d --build` command. |
| `redirect_uri_mismatch` | Add the exact localhost callback URI above to the Google OAuth client. |
| Finish now asks you to sign in | Sign in again **in the browser profile that owns the room timer**. A separate admin sign-in can replace the cookie if it uses the same profile. |
| Finish now says the account does not own the timer | The current cookie belongs to a different user. Return to the room participant's browser profile. |
| Finish now says no running focus block | Refresh the room and start a new work timer. A discarded cycle cannot earn a reward. |
| Docker reports a port in use | Stop the host service using that port, especially a local MongoDB server on 27017. |

To stop containers **without deleting demo data**:

```bash
docker compose -f docker-compose.yml -f docker-compose.dev.yml --profile app down
```

Avoid `down -v` or `pnpm docker:reset` before the demo; those delete the
database volumes.
