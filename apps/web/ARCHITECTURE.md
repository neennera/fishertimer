# Web Application Architecture (Next.js 16)

## Architecture Pattern: Feature-Sliced Modular Architecture

The frontend is structured to scale cleanly as multiple microservices and domain features expand:

```
apps/web/
├── app/                           # Next.js App Router (pages, layouts, routes)
│   ├── layout.tsx                 # Root layout with global theme & fonts
│   └── page.tsx                   # Main dashboard landing page
├── components/                    # Global shared UI components
│   └── ui/                        # Buttons, inputs, modals, cards
├── features/                      # Feature-sliced modules
│   ├── auth/                      # Google OAuth sign-in & session state
│   ├── timer/                     # Study timer widget & controls
│   ├── session/                   # Room browser, creation modal, rosters
│   ├── rewards/                   # FishTank inventory & drop animations
│   ├── leaderboard/               # Ranking tables & filters
│   └── admin/                     # Moderation tools & live session inspection
├── hooks/                         # Global React hooks
├── lib/                           # Utilities & typed API clients
│   └── api-client.ts              # Centralized service fetcher
└── types/                         # Local view models
```

## Rules for Frontend Code
1. Place domain-specific UI and logic into `features/<name>/`.
2. Keep general reusable UI elements in `@repo/ui` or `components/ui/`.
3. Use `fetchFromService()` in `lib/api-client.ts` to communicate with backend microservices.
