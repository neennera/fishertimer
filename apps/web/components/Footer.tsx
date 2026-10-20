// The group members listed in the project README.
const MAKERS = [
  "Chanatda Konchom",
  "Ittichet Thongsang",
  "Napat Srisamut",
  "Naphat Serirak",
  "Chayut Archamongkol",
];

/**
 * Bottom of a content page: who made it, in two small lines. It sits straight on
 * the scene (no panel), so the shadow keeps the white text readable on water.
 */
export function Footer() {
  return (
    <footer className="mx-auto w-full max-w-5xl px-4 pb-6 pt-2 text-center text-sm leading-snug text-white [text-shadow:0_1px_0_var(--color-ink)]">
      <p>© 2026 Fisher Timer. Made by {MAKERS.join(", ")}.</p>
      <p>Software Architecture project.</p>
    </footer>
  );
}
