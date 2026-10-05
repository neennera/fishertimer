import Link from "next/link";
import { SessionHeader } from "../components/SessionHeader";
import { ParallaxScene } from "../components/ui/ParallaxScene";
import { PixelButton } from "../components/ui/PixelButton";
import { PixelPanel } from "../components/ui/PixelPanel";
import { SIGNIN_SCENE_LAYERS } from "../lib/scenes/signin-scene";

const FEATURES = [
  {
    emoji: "⏱️",
    title: "Study Timer",
    description: "Focus and rest cycles to keep you on track",
    href: "/timer",
  },
  {
    emoji: "👥",
    title: "Study Sessions",
    description: "Join and create study rooms with peers",
    href: null, // no page yet
  },
  {
    emoji: "🐟",
    title: "FishTank Rewards",
    description: "Earn fish by finishing focus sessions",
    href: "/account",
  },
  {
    emoji: "🏆",
    title: "Leaderboard",
    description: "See how your fish score ranks",
    href: "/leaderboard",
  },
];

const CARD =
  "pixel-panel flex h-full flex-col gap-1 p-4 text-left transition-transform duration-75";

/** Home: the scene background plus one clickable card per feature. */
export default function Home() {
  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="mx-auto w-full max-w-3xl px-4 py-10 sm:py-14">
        <PixelPanel>
          <h1
            className="font-display leading-tight"
            style={{ fontSize: "calc(var(--px) * 10)" }}
          >
            🎣 Fisher Timer
          </h1>
          <p className="mt-2 text-bark">
            Study in focused sessions, catch fish, climb the leaderboard.
          </p>

          <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
            {FEATURES.map(({ emoji, title, description, href }) => {
              const body = (
                <>
                  <span className="text-3xl" aria-hidden="true">
                    {emoji}
                  </span>
                  <h2 className="font-display">{title}</h2>
                  <p className="text-bark">{description}</p>
                  <span className="mt-auto pt-2 font-label text-xs text-bark">
                    {href ? "Open →" : "Coming soon"}
                  </span>
                </>
              );
              return href ? (
                <Link
                  key={title}
                  href={href}
                  className={`${CARD} hover:-translate-y-0.5 focus-visible:-translate-y-0.5`}
                  style={{ background: "var(--color-cream-2)" }}
                >
                  {body}
                </Link>
              ) : (
                <div
                  key={title}
                  className={`${CARD} opacity-60`}
                  style={{ background: "var(--color-cream-2)" }}
                  aria-disabled="true"
                >
                  {body}
                </div>
              );
            })}
          </div>

          <div className="mt-6">
            <Link href="/timer">
              <PixelButton>Start Studying</PixelButton>
            </Link>
          </div>
        </PixelPanel>
      </main>
    </div>
  );
}
