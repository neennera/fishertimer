"use client";

import { Suspense, useEffect, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelInput } from "../../../components/ui/PixelInput";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { completeFirstTimeSetup, getSession, handleAuthCallback, type Session } from "../../../lib/auth";
import { validateDisplayName } from "../../../lib/validate-display-name";

function WelcomeForm() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const [session, setSession] = useState<Session | null>(null);
  const [name, setName] = useState("");
  const [attemptedSubmit, setAttemptedSubmit] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      let current = await getSession();

      // No callback route exists yet to land here for real, so there's no
      // session in this tab when testing directly — ?mockScenario=
      // setup-default/setup-name-empty/setup-name-too-long lets this page
      // be reached directly for each state, the same way /signin's
      // ?error= links work.
      if (!current) {
        const result = await handleAuthCallback(searchParams);
        current = result.ok ? result.session : null;
      }

      if (cancelled) {
        return;
      }

      if (!current) {
        router.replace("/signin");
        return;
      }

      setSession(current);
      setName(current.user.displayName);
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, [router, searchParams]);

  const liveError = validateDisplayName(name);
  const isEmpty = name.trim().length === 0;
  // "Display name can't be empty" only shows once they've tried to
  // continue — "too long" shows live, as they type past the limit (02b vs
  // 02c in the wireframes).
  const displayedError = attemptedSubmit || !isEmpty ? liveError : undefined;

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setAttemptedSubmit(true);
    setSubmitError(null);

    if (validateDisplayName(name)) {
      return;
    }

    setSubmitting(true);
    try {
      const result = await completeFirstTimeSetup(name);
      if (!result.ok) {
        return;
      }
      router.push("/");
    } catch {
      setSubmitError("Something went wrong. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  // Panel renders immediately; only the session-dependent fields (name,
  // email) show a skeleton while they load.
  return (
    <PixelPanel className="w-full max-w-sm text-center">
      <h1 className="font-display text-3xl leading-none">Welcome to Fisher Timer</h1>
      <p className="mt-3 text-sm text-bark">
        Let&rsquo;s set up your profile before you start
      </p>

      <form className="mt-6 text-left" onSubmit={handleSubmit} aria-busy={!session}>
        {session ? (
          <PixelInput
            label="Display name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            aria-invalid={displayedError ? "true" : undefined}
            maxLength={100}
          />
        ) : (
          <div>
            <span className="pixel-label">Display name</span>
            <div className="pixel-input pixel-skeleton" aria-hidden="true" />
          </div>
        )}

        <p className="pixel-field-error" role={displayedError || submitError ? "alert" : undefined}>
          {displayedError ?? submitError ?? " "}
        </p>

        {session ? (
          <p className="text-sm text-bark">
            Signed in as {session.user.email}
          </p>
        ) : (
          <p className="mt-4 text-sm text-bark" aria-hidden="true">
            <span className="pixel-skeleton pixel-skeleton--text" />
          </p>
        )}

        <PixelButton block type="submit" className="mt-2" disabled={submitting || !session}>
          {submitting ? "Saving…" : "Continue"}
        </PixelButton>
      </form>
    </PixelPanel>
  );
}

export default function WelcomePage() {
  return (
    <Suspense fallback={null}>
      <WelcomeForm />
    </Suspense>
  );
}
