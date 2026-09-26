# Account Service

## Overview
Google OAuth authentication (UC-06): sign in, sign up on first login, sign out.

## Port
Default port: `8082` (the browser reaches it through the API Gateway as `/api/auth/*` or `/api/account/*`).

## HTTP API

| Method | Gateway route | Service route | Purpose |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/auth/google/login` | `/api/v1/account/google/login` | Start sign-in. Optional `?next=/path` (where to land when signed in, default `/`) and `?signup=/path` (where to land when the e-mail has no account, default `/signup`); both must be same-site paths. Sets the `state` cookie and redirects to Google. |
| `GET` | `/api/auth/google/callback` | `/api/v1/account/google/callback` | Google redirect target: checks `state`, exchanges the code, then asks whether the e-mail exists. Yes → `ft_session` + redirect to `next`. No → 15-min `ft_signup` ticket, no row written, redirect to `signup`. |
| `GET` | `/api/auth/me` | `/api/v1/account/me` | Returns the signed-in user (`200` + user JSON), or `401 {"error":"sign in required"}` if there is no valid session. Reads the session from the `ft_session` cookie or an `Authorization: Bearer <token>` header. |
| `POST` | `/api/auth/signup` | `/api/v1/account/signup` | `{display_name}` → creates the account for the ticket's e-mail and signs the user in (`201`). |
| `POST` | `/api/auth/signout` | `/api/v1/account/signout` | Clears the cookies. |
| `PATCH` | `/api/auth/update-profile` | `/api/v1/account/update-profile` | `{display_name}` → renames the signed-in user and re-issues `ft_session` (`200` + updated user). Requires a session (`ft_session` cookie or `Authorization: Bearer <token>`). Errors: `400` (missing/too-long display name), `401` (no/expired session). |
| `GET` | `/api/auth/profile?id=<user_id>` | `/api/v1/account/profile?id=<user_id>` | Looks up any user by id (no session required) → `200` + user JSON. Errors: `400` (missing `id`), `404` (no such user). |
| `GET` | - | `/health`, `/api/v1/account/status` | Liveness and configuration status. |

## Environment

| Variable | Purpose |
| :--- | :--- |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | OAuth client from Google Cloud Console. |
| `GOOGLE_REDIRECT_URI` | Must match the registered redirect URI exactly (`http://localhost:3000/api/auth/google/callback` locally). |
| `JWT_SECRET` / `JWT_ISSUER` / `JWT_EXPIRATION_HOURS` | Session token signing (`168` = the 7 days UC-06 specifies). |
| `ACCOUNT_DATABASE_URL` | `account_db` connection string. **Required** - the service exits if the database is unreachable. |
| `NEXT_PUBLIC_APP_URL` | Where the user is redirected after sign-in. |

Values come from the process environment, falling back to the monorepo root `.env`.

## Database
`account_db` (PostgreSQL, local port `5432`), table `users` —
see [`database/schemas/001_create_users_table.sql`](database/schemas/001_create_users_table.sql).

## Scripts
- `pnpm dev` : Runs the service locally with Go
- `pnpm build` : Compiles the Go binary to `bin/server`
- `pnpm test` : Runs Go tests
- `pnpm lint` : Runs Go static analysis (`go vet`)
