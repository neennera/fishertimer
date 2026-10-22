"use client";

import { useState } from "react";
import Link from "next/link";
import { Logout } from "pixelarticons/react/Logout";
import { cx } from "../lib/cx";
import { initials as initialsOf } from "../lib/initials";
import { PixelButton } from "./ui/PixelButton";

// Fish shown beside the wordmark (apps/web/public/sprites/fish).
const BRAND_FISH = ["/sprites/fish/Goldfish.png", "/sprites/fish/Clownfish.png", "/sprites/fish/Koi.png"];

export interface HeaderProps {
  /** Omitted while signed out. The bar keeps the same height either way. */
  user?: { displayName: string; avatarUrl?: string; isAdmin?: boolean } | null;
  /** Shows the sign-out button (only alongside `user`). */
  onSignOut?: () => void;
  /** Disables the sign-out button while the request is in flight. */
  signingOut?: boolean;
  /** Only once known signed out, so it never flashes while loading. */
  showSignIn?: boolean;
  /** The sign-in check failed (offline, timeout): shows Retry instead of Sign in. */
  onRetrySession?: () => void;
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
  onRetrySession,
  className,
}: HeaderProps) {
  // Keyed by URL, so a new avatar gets a fresh attempt.
  const [failedAvatar, setFailedAvatar] = useState<string | null>(null);
  const avatarUrl = user?.avatarUrl && failedAvatar !== user.avatarUrl ? user.avatarUrl : null;

  const initials = user?.displayName ? initialsOf(user.displayName) : null;

  return (
    <header className={cx("pixel-header", className)}>
      <Link href="/" className="pixel-header__brand">
        <span aria-hidden="true" className="flex items-center -space-x-1">
          {BRAND_FISH.map((src, i) => (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              key={src}
              src={src}
              alt=""
              width={28}
              height={28}
              className={cx("pixel-sprite", i > 0 && "hidden sm:block")}
              style={{
                imageRendering: "pixelated",
                width: 28,
                height: 28,
                transform: i % 2 ? "scaleX(-1)" : undefined,
              }}
            />
          ))}
        </span>
        <span className="pixel-header__title">Fisher Timer</span>
      </Link>

      {user?.isAdmin && (
        <nav
          aria-label="Developer pages"
          className="ml-auto flex items-center gap-1 font-label sm:gap-3"
          style={{ fontSize: "calc(var(--px) * 3.6)" }}
        >
          <Link href="/admin" className="inline-flex min-h-11 items-center whitespace-nowrap px-1 underline-offset-4 hover:underline focus-visible:underline">
            admin page
          </Link>
          <Link href="/dev" className="inline-flex min-h-11 items-center whitespace-nowrap px-1 underline-offset-4 hover:underline focus-visible:underline">
            dev page
          </Link>
        </nav>
      )}


      {user && initials && (
        <Link
          href="/account"
          className="pixel-btn pixel-btn--ghost pixel-btn--icon pixel-header__avatar"
          // The admin links already push everything right; without them the avatar does.
          style={user.isAdmin ? { marginLeft: 0 } : undefined}
          title={`${user.displayName} — your account`}
          aria-label="Your account"
        >
          <span className="pixel-header__avatar-face">
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
          </span>
        </Link>
      )}

      {!user && showSignIn && (
        <Link href="/signin" className="pixel-btn pixel-btn--sm pixel-header__action whitespace-nowrap">
          Sign in
        </Link>
      )}

      {!user && onRetrySession && (
        <PixelButton
          small
          className="pixel-header__action min-h-11 sm:min-h-0"
          onClick={onRetrySession}
          aria-label="Retry sign-in check"
          title="Couldn't check your sign-in. Try again."
        >
          Retry
        </PixelButton>
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
