import Link from "next/link";
import { ArrowRight } from "pixelarticons/react/ArrowRight";
import { Footer } from "../components/Footer";
import { HangingSign } from "../components/HangingSign";
import { SessionHeader } from "../components/SessionHeader";
import { ParallaxScene } from "../components/ui/ParallaxScene";
import { PixelPanel } from "../components/ui/PixelPanel";
import { SIGNIN_SCENE_LAYERS } from "../lib/scenes/signin-scene";

// The three steps of a session, in order.
const STEPS = [
  { title: "Join a room", description: "Pick an open room or start your own." },
  { title: "Focus, then rest", description: "Work in timed blocks with breaks." },
  { title: "Catch a fish", description: "Every finished block earns a fish." },
];

// Where to go next. The timer itself lives inside a room, so "Study rooms"
// is also the way to the timer.
const DESTINATIONS = [
  {
    sprite: "/sprites/ui/icon-room.png",
    title: "Study rooms",
    description: "Join or create a room and use the focus timer with others.",
    href: "/rooms",
  },
  {
    sprite: "/sprites/ui/trophy.png",
    title: "Leaderboard",
    description: "See how your fish score ranks this week, month or all time.",
    href: "/leaderboard",
  },
  {
    sprite: "/sprites/fish/Goldfish.png",
    title: "My fish",
    description: "Your catch, stats and account.",
    href: "/account",
  },
];

/**
 * Home: one clear action first (Start Studying), how it works, then the places
 * to go. Text sits on parchment panels, never straight on the scene.
 */
export default function Home() {
  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <SessionHeader />
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 pb-8 pt-44 sm:pb-12">
        {/* The sign stands on the top edge of this panel and carries the title. */}
        <div className="relative">
          <HangingSign />
          <PixelPanel as="section" aria-labelledby="home-heading">
            <h1 id="home-heading" className="sr-only">
              Fisher Timer
            </h1>
            <p className="max-w-prose text-lg text-ink">
              Study in focused sessions, catch fish, climb the leaderboard.
            </p>

            <div className="mt-6 flex flex-wrap items-center gap-3">
              <Link href="/rooms" className="pixel-btn min-h-11">
                Start Studying
              </Link>
              <Link href="/settings" className="pixel-btn pixel-btn--ghost min-h-11">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src="/sprites/ui/gear.png" alt="" className="pixel-sprite h-6 w-6" />
                Timer defaults
              </Link>
            </div>
          </PixelPanel>
        </div>

        <PixelPanel as="section" aria-labelledby="how-heading">
          <h2 id="how-heading" className="font-display text-2xl">
            How it works
          </h2>
          <ol className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
            {STEPS.map(({ title, description }, i) => (
              <li
                key={title}
                className="flex flex-col gap-1 p-3 sm:p-4"
                style={{ background: "var(--color-tile)" }}
              >
                <span className="font-label text-sm text-bark">Step {i + 1}</span>
                <span className="font-display text-xl">{title}</span>
                <span className="text-base text-ink">{description}</span>
              </li>
            ))}
          </ol>
        </PixelPanel>

        <PixelPanel as="nav" aria-labelledby="go-heading">
          <h2 id="go-heading" className="font-display text-2xl">
            Go to
          </h2>
          <ul className="mt-4 flex flex-col gap-3">
            {DESTINATIONS.map(({ sprite, title, description, href }) => (
              <li key={title}>
                <Link
                  href={href}
                  className="group flex min-h-16 items-center gap-4 p-4 transition-transform duration-75 hover:-translate-y-0.5 focus-visible:-translate-y-0.5"
                  style={{ background: "var(--color-tile)" }}
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={sprite} alt="" className="pixel-sprite h-12 w-12 flex-none" />
                  <span className="flex min-w-0 flex-1 flex-col">
                    <span className="font-display text-xl">{title}</span>
                    <span className="text-base text-ink">{description}</span>
                  </span>
                  <ArrowRight aria-hidden="true" className="h-6 w-6 flex-none" />
                </Link>
              </li>
            ))}
          </ul>
        </PixelPanel>
      </main>
      <div className="mt-auto" />
      <Footer />
    </div>
  );
}
