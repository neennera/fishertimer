// Characters that show nothing: soft hyphen, zero-width space / non-joiner /
// joiner, direction marks, line and paragraph separators, direction overrides,
// word joiner and invisible operators, BOM. Built from code points so the source
// holds no invisible characters.
const INVISIBLE_RANGES: Array<[number, number]> = [
  [0x00ad, 0x00ad],
  [0x200b, 0x200f],
  [0x2028, 0x202e],
  [0x2060, 0x2064],
  [0xfeff, 0xfeff],
];
const invisibleClass = INVISIBLE_RANGES.map(([from, to]) =>
  from === to
    ? String.fromCharCode(from)
    : `${String.fromCharCode(from)}-${String.fromCharCode(to)}`,
).join("");
// Only at the start of a word: a zero-width joiner inside "👨‍👩‍👧" must stay.
const LEADING_INVISIBLE = new RegExp(`^[${invisibleClass}]+`);

// A grapheme, not a UTF-16 unit or code point: "🐟" and flags like "🇺🇸" stay whole.
const segmenter =
  typeof Intl !== "undefined" && "Segmenter" in Intl
    ? new Intl.Segmenter(undefined, { granularity: "grapheme" })
    : null;

function firstCharacter(word: string): string {
  if (segmenter) {
    for (const { segment } of segmenter.segment(word)) {
      return segment;
    }
    return "";
  }
  return Array.from(word)[0] ?? "";
}

function visibleWords(name: string): string[] {
  return name
    .trim()
    .split(/\s+/)
    .map((word) => word.replace(LEADING_INVISIBLE, ""))
    .filter(Boolean);
}

/** False for a name that is empty or made only of spaces and invisible characters. */
export function hasVisibleText(name: string): boolean {
  return visibleWords(name).length > 0;
}

/**
 * Up to two initials of a display name, by character; invisible characters never
 * become an initial. "?" when the name has no visible characters, so an avatar
 * (and the header's account link, which needs something to show) is never blank.
 */
export function initials(name: string): string {
  const letters = visibleWords(name).map(firstCharacter).slice(0, 2).join("").toUpperCase();
  return letters || "?";
}
