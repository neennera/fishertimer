# Web Application Architecture (Next.js 16)

## Architecture Pattern: Feature-Sliced Modular Architecture

The frontend is structured to scale cleanly as multiple microservices and domain features expand:

```
apps/web/
├── app/                           # Next.js App Router (pages, layouts, routes)
│   ├── layout.tsx                 # Root layout: font variables + globals
│   ├── page.tsx                   # Landing; redirects to /account or /signin once auth lands
│   ├── globals.css                # Tailwind import + base only. Keep it small
│   ├── tokens.css                 # DESIGN TOKENS — the only file with raw hex or px
│   ├── pixel.css                  # Pixel surface primitives, all derived from --px
│   ├── fonts.ts                   # next/font/google, mapped to semantic type roles
│   └── styleguide/                # Every component and state, built from the real components
├── components/                    # Global shared UI
│   ├── Header.tsx                 # The walnut bar on every screen
│   └── ui/                        # PixelButton, PixelPanel, PixelInput, PixelAlert,
│                                  # PixelBadge, StatTile
├── features/                      # Feature-sliced modules
│   ├── auth/                      # Google OAuth sign-in & session state (UC-06)
│   ├── account/                   # Profile, display name, stats dashboard (UC-07)
│   ├── timer/                     # Study timer widget & controls
│   ├── session/                   # Room browser, creation modal, rosters
│   ├── rewards/                   # FishTank inventory & drop animations
│   ├── leaderboard/               # Ranking tables & filters
│   └── admin/                     # Moderation tools & live session inspection
├── hooks/                         # Global React hooks
├── lib/                           # Utilities & typed API clients
│   ├── api-client.ts              # Centralized service fetcher
│   └── cx.ts                      # Class-name joiner
├── mocks/                         # Fixture data while the Go services are stubs
└── types/                         # Local view models
```

## Rules for Frontend Code

1. Place domain-specific UI and logic into `features/<name>/`.
2. Keep general reusable UI elements in `@repo/ui` or `components/ui/`.
3. Use `fetchFromService()` in `lib/api-client.ts` to communicate with backend microservices.
4. **Components never call `fetch` and never import from `mocks/`.** They call a
   function in the feature's `*.api.ts`, which is the only place that knows
   whether data is real. Switched by `NEXT_PUBLIC_USE_MOCKS`.
5. **No raw hex colour and no raw px value outside `app/tokens.css`.** Every
   border, bevel, depth band, focus ring and sprite scale is a multiple of
   `--px`. See [`DESIGN_SYSTEM.md`](./DESIGN_SYSTEM.md).
6. **Tailwind for layout, `pixel.css` for surfaces.** Utilities handle flex,
   grid, gap and breakpoints; pixel geometry stays in the component layer.

## Styling

TailwindCSS v4 (ADR-002), configured CSS-first — there is no `tailwind.config.js`.
The theme lives in the `@theme` block of `app/tokens.css`, so every colour and
type role is available both as a utility (`bg-amber`, `font-display`) and as a
custom property (`var(--color-amber)`).

The full design system, including the art-pixel grid and the rules that keep
this a productivity tool rather than a game, is documented in
[`DESIGN_SYSTEM.md`](./DESIGN_SYSTEM.md). Read it before adding a component.

`/styleguide` renders every component in every state. If a change breaks the
system, it breaks there first.
