/**
 * Minimal class-name joiner. Deliberately not `clsx` — this is the only
 * behaviour we need and it keeps the dependency list honest.
 */
export function cx(...parts: Array<string | false | null | undefined>): string {
  return parts.filter(Boolean).join(" ");
}
