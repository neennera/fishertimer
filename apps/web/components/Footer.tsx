import Link from "next/link";
import { PixelPanel } from "./ui/PixelPanel";

// The group members listed in the project README.
const MAKERS = [
  "Chanatda Konchom",
  "Ittichet Thongsang",
  "Napat Srisamut",
  "Naphat Serirak",
  "Chayut Archamongkol",
];

const LINKS = [
  { href: "/rooms", label: "Study rooms" },
  { href: "/leaderboard", label: "Leaderboard" },
  { href: "/account", label: "My fish" },
  { href: "/settings", label: "Timer defaults" },
];

/** Bottom of a content page: where to go, and who made it. */
export function Footer() {
  return (
    <footer className="mx-auto w-full max-w-3xl px-4 pb-8">
      <PixelPanel>
        <nav aria-label="Footer">
          <ul className="flex flex-wrap gap-x-2 gap-y-1">
            {LINKS.map(({ href, label }) => (
              <li key={href}>
                <Link
                  href={href}
                  className="inline-flex min-h-11 items-center px-2 text-base underline underline-offset-4 hover:text-bark"
                >
                  {label}
                </Link>
              </li>
            ))}
          </ul>
        </nav>
        <p className="mt-3 text-base text-ink">
          © 2026 Fisher Timer. Made by {MAKERS.join(", ")}.
        </p>
        <p className="mt-1 text-sm text-bark">Software Architecture project.</p>
      </PixelPanel>
    </footer>
  );
}
