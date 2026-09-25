import Link from "next/link";
import { SessionHeader } from "../components/SessionHeader";
import { PixelButton } from "../components/ui/PixelButton";
import { PixelPanel } from "../components/ui/PixelPanel";

const FEATURES = [
  {
    emoji: "⏱️",
    title: "Study Timer",
    description: "Focus & rest cycles synced via API Gateway",
  },
  {
    emoji: "👥",
    title: "Study Sessions",
    description: "Join & create study rooms with peers",
  },
  {
    emoji: "🐟",
    title: "FishTank Rewards",
    description: "Earn fish rewards upon ending sessions",
  },
  {
    emoji: "🏆",
    title: "Leaderboard",
    description: "Weekly & all-time focus rankings",
  },
];

/**
 * Placeholder home. Once the auth branch lands this redirects to /account
 * when a session exists and /signin when it doesn't (UC-06).
 */
export default function Home() {
  return (
    <>
      <SessionHeader />
      <main className="mx-auto w-full max-w-3xl px-4 py-12">
        <PixelPanel>
          <h1
            className="font-display leading-tight"
            style={{ fontSize: "calc(var(--px) * 10)" }}
          >
            🎣 Fisher Timer — Student Portal
          </h1>
          <p className="mt-2 text-bark">
            Frontend Website (Client page) — communicating via API Gateway
            (Port 8080).
          </p>

          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
            {FEATURES.map(({ emoji, title, description }) => (
              <PixelPanel as="article" key={title}>
                <h3 className="font-display">
                  {emoji} {title}
                </h3>
                <p className="mt-1 text-bark">{description}</p>
              </PixelPanel>
            ))}
          </div>

          <div className="mt-6 flex flex-wrap items-center gap-4">
            <PixelButton>Start Studying</PixelButton>
            <Link href="/admin" className="underline underline-offset-4">
              🛡️ Switch to Admin Portal
            </Link>
            <Link href="/styleguide" className="underline underline-offset-4">
              View the styleguide
            </Link>
          </div>
        </PixelPanel>
      </main>
    </>
  );
}
