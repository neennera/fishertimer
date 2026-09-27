"use client";

import { useState } from "react";
import Link from "next/link";
import { Logout } from "pixelarticons/react/Logout";
import { cx } from "../lib/cx";
import { PixelButton } from "./ui/PixelButton";

export interface HeaderProps {
  /** Omitted while signed out. The bar keeps the same height either way. */
  user?: { displayName: string; avatarUrl?: string } | null;
  /** Shows the sign-out button (only alongside `user`). */
  onSignOut?: () => void;
  /** Disables the sign-out button while the request is in flight. */
  signingOut?: boolean;
  /** Only once known signed out, so it never flashes while loading. */
  showSignIn?: boolean;
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
export function Header({
  user,
  onSignOut,
  signingOut = false,
  showSignIn = false,
  className,
}: HeaderProps) {
  // Keyed by URL, so a new avatar gets a fresh attempt.
  const [failedAvatar, setFailedAvatar] = useState<string | null>(null);
  const avatarUrl = user?.avatarUrl && failedAvatar !== user.avatarUrl ? user.avatarUrl : null;

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
        <Link
          href="/account"
          className="pixel-header__avatar"
          title={`${user.displayName} — your account`}
          aria-label="Your account"
        >
          {avatarUrl ? (
            // A remote Google photo — see the profile panel's avatar.
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={avatarUrl}
              alt=""
              className="pixel-header__avatar-img"
              referrerPolicy="no-referrer"
              onError={() => setFailedAvatar(avatarUrl)}
            />
          ) : (
            initials
          )}
        </Link>
      )}

      {!user && showSignIn && (
        <Link href="/signin" className="pixel-btn pixel-btn--sm pixel-header__action">
          Sign in
        </Link>
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
