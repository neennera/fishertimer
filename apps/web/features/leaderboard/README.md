# Feature: Leaderboard (UC-08)

The Leaderboard feature enables users to view rankings of anglers based on fish rewards earned across **Weekly**, **Monthly**, and **All-Time** periods.

---

## Directory Structure

```
features/leaderboard/
├── components/
│   ├── PeriodTabs.tsx         # Responsive 3-column period filter tab strip (S-2)
│   └── LeaderboardTable.tsx   # Ranked list with podium badges, user avatar, and YOU highlight (NF-6)
├── hooks/
│   └── useLeaderboard.ts      # Custom hook handling S-1 cache fetch, auto-refetch, and manual refresh
├── types/
│   └── index.ts               # TypeScript interfaces matching backend domain models
└── README.md
```

---

## Routes & Pages

- **`/leaderboard`**: Main leaderboard view ([`app/leaderboard/page.tsx`](../../app/leaderboard/page.tsx))
  - Signed-in only (redirects to `/signin`); scene background shared with the auth pages.
  - Summary tiles: Current Leader and Your Ranking.
  - Compact period tabs (`weekly`, `monthly`, `all-time`; default `monthly`) on the same row as Refresh.
  - Table: 5 rows per page, always-visible `<` / `>` pager, ranks 1-3 share one highlight, viewer's row in blue when on the page, and the viewer's position (or "Unranked") pinned under a divider.
  - Skeleton loading states and empty period handling.

---

## Design System Integration

All styling strictly follows the Fisher Timer pixel design system defined in `app/tokens.css` and `app/pixel.css`:
- **Panels & Surfaces**: `.pixel-panel` with `--pixclip` bevel.
- **Buttons**: `.pixel-btn` (primary amber) and `.pixel-btn--ghost` (neutral).
- **Highlights**: Lake blue (`--color-lake`) for active user row.
- **Podiums**: Distinct gold/amber (`#1`), silver (`#2`), and bronze (`#3`) badge accents.
