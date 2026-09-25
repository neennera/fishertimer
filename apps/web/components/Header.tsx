import Link from "next/link";
import { Logout } from "pixelarticons/react/Logout";
import { cx } from "../lib/cx";
import { PixelButton } from "./ui/PixelButton";

export interface HeaderProps {
  /** Omitted while signed out. The bar keeps the same height either way. */
  user?: { displayName: string } | null;
  /** Shows the sign-out button (only alongside `user`). */
  onSignOut?: () => void;
  /** Disables the sign-out button while the request is in flight. */
  signingOut?: boolean;
  className?: string;
}

/**
 * The bar across the top of every page.
 *
 * Height is fixed in `.pixel-header` so it does not change between pages.
 * The logo is a placeholder block until the 32x32 sprite exists — replace the
 * span with <img className="pixel-sprite"> then.
 *
 * Presentational only: it never loads the session itself. SessionHeader wires
 * it to lib/auth.ts.
 */
export function Header({ user, onSignOut, signingOut = false, className }: HeaderProps) {
  const initials = user?.displayName
    ? user.displayName
        .split(" ")
        .map((part) => part[0])
        .slice(0, 2)
        .join("")
        .toUpperCase()
    : null;

  return (
    <header className={cx("pixel-header", className)}>
      <Link href="/" className="pixel-header__brand">
        <span aria-hidden="true" className="pixel-header__logo" />
        <span className="pixel-header__title">Fisher Timer</span>
      </Link>

      {user && initials && (
        <span className="pixel-header__avatar" title={user.displayName}>
          {initials}
        </span>
      )}

      {user && onSignOut && (
        <PixelButton
          variant="danger"
          icon
          className="pixel-header__action"
          onClick={onSignOut}
          disabled={signingOut}
          aria-label="Sign out"
          title="Sign out"
        >
          <Logout aria-hidden="true" />
        </PixelButton>
      )}
    </header>
  );
}
