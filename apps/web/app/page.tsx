import Link from "next/link";
import { ArrowRight } from "pixelarticons/react/ArrowRight";
import { Fish } from "pixelarticons/react/Fish";
import { Gear } from "pixelarticons/react/Gear";
import { Trophy } from "pixelarticons/react/Trophy";
import { Users } from "pixelarticons/react/Users";
import { SessionHeader } from "../components/SessionHeader";
import { ParallaxScene } from "../components/ui/ParallaxScene";
import { PixelPanel } from "../components/ui/PixelPanel";
import { SIGNIN_SCENE_LAYERS } from "../lib/scenes/signin-scene";

// Where to go next. The timer itself lives inside a room, so "Study rooms"
// is also the way to the timer.
const DESTINATIONS = [
  {
    icon: Users,
    title: "Study rooms",
    description: "Join or create a room and use the focus timer with others.",
    href: "/rooms",
  },
  {
    icon: Trophy,
    title: "Leaderboard",
    description: "See how your fish score ranks this week, month or all time.",
    href: "/leaderboard",
  },
  {
    icon: Fish,
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
      <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-8 sm:py-12">
        <PixelPanel>
          <h1
            className="font-display leading-tight"
            style={{ fontSize: "calc(var(--px) * 10)" }}
          >
            <Fish aria-hidden="true" className="mr-2 inline-block h-[1em] w-[1em] align-[-0.1em]" />
            Fisher Timer
          </h1>
          <p className="mt-2 text-bark">
            Study in focused sessions, catch fish, climb the leaderboard.
          </p>

          <div className="mt-6 flex flex-wrap items-center gap-4">
            <Link href="/rooms" className="pixel-btn">
              Start Studying
            </Link>
            <Link href="/settings" className="underline underline-offset-4">
              <Gear aria-hidden="true" className="mr-1 inline-block h-[1.1em] w-[1.1em] align-[-0.15em]" />
              Timer defaults
            </Link>
          </div>
        </PixelPanel>

        <PixelPanel as="nav" aria-labelledby="go-heading">
          <h2 id="go-heading" className="font-display text-2xl">
            Go to
          </h2>
          <ul className="mt-4 flex flex-col gap-3">
            {DESTINATIONS.map(({ icon: Icon, title, description, href }) => (
              <li key={title}>
                <Link
                  href={href}
                  className="group flex min-h-16 items-center gap-4 p-4 transition-transform duration-75 hover:-translate-y-0.5 focus-visible:-translate-y-0.5"
                  style={{ background: "var(--color-tile)" }}
                >
                  <Icon aria-hidden="true" className="h-8 w-8 flex-none" />
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
    </div>
  );
}
