"use client";

import { Suspense, useEffect, useRef, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelCheckbox } from "../../../components/ui/PixelCheckbox";
import { PixelInput } from "../../../components/ui/PixelInput";
import { PixelModal } from "../../../components/ui/PixelModal";
import { PixelPanel } from "../../../components/ui/PixelPanel";
import { AUTH_ERROR_PARAM, completeFirstTimeSetup, getPendingSignUp, getSession } from "../../../lib/auth";
import { validateDisplayName } from "../../../lib/validate-display-name";

function WelcomeForm() {
  const router = useRouter();
  const [name, setName] = useState("");
  // Set once the user types: a late suggested name must not overwrite their own.
  const typed = useRef(false);
  const [attemptedSubmit, setAttemptedSubmit] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [consentChecked, setConsentChecked] = useState(false);
  const [privacyOpen, setPrivacyOpen] = useState(false);

  // Someone already signed in has nothing to set up: go to the account page. Otherwise
  // read the pending sign-up (GET /me can't say whether one is pending; that is only in
  // the ft_signup cookie). Its Google name starts the display name; the user can change
  // it, and a name over the limit shows its error right away. No ticket (or an expired
  // one) sends them back to sign in; a failed lookup just leaves the field empty. The two
  // calls are chained so a signed-in user is redirected once, not twice. In mock mode,
  // /welcome?mockScenario=setup-* reaches each 02 state directly.
  useEffect(() => {
    let cancelled = false;
    void (async () => {
      const session = await getSession();
      if (cancelled) {
        return;
      }
      if (session.status === "signed_in") {
        router.replace("/account");
        return;
      }
      const pending = await getPendingSignUp();
      if (cancelled) {
        return;
      }
      if (pending.status === "expired") {
        router.replace(`/signin?${AUTH_ERROR_PARAM}=signup_expired`);
      } else if (pending.status === "ready" && !typed.current) {
        setName(pending.suggestedName);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [router]);

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

    if (validateDisplayName(name) || !consentChecked) {
      return;
    }

    setSubmitting(true);
    try {
      const result = await completeFirstTimeSetup(name);
      if (result.ok) {
        router.push("/account");
      } else if (result.code === "invalid") {
        setSubmitError(result.error);
      } else {
        // account_creation_failed (01d) or signup_expired: both are shown on
        // the sign-in screen, where they can start again.
        router.replace(`/signin?${AUTH_ERROR_PARAM}=${result.code}`);
      }
    } catch {
      setSubmitError("Something went wrong. Please try again.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <PixelPanel className="w-full max-w-sm text-center">
      <h1 className="font-display text-3xl leading-none">Welcome to Fisher Timer</h1>
      <p className="mt-3 text-sm text-bark">
        Let&rsquo;s set up your profile before you start
      </p>

      <form className="mt-6 text-left" onSubmit={handleSubmit}>
        <PixelInput
          label="Display name"
          value={name}
          onChange={(event) => {
            typed.current = true;
            setName(event.target.value);
          }}
          aria-invalid={displayedError ? "true" : undefined}
        />

        <p className="pixel-field-error" role={displayedError || submitError ? "alert" : undefined}>
          {displayedError ?? submitError ?? " "}
        </p>

        <PixelCheckbox
          className="mt-4"
          checked={consentChecked}
          onChange={(event) => setConsentChecked(event.target.checked)}
          required
        >
          I have read and acknowledge the{" "}
          <button type="button" className="pixel-link" onClick={() => setPrivacyOpen(true)}>
            Privacy Notice
          </button>{" "}
          and consent to the collection and use of my personal data as described above.
        </PixelCheckbox>

        <PixelButton
          block
          type="submit"
          className="mt-4"
          disabled={submitting || !consentChecked}
        >
          {submitting ? "Saving…" : "Continue"}
        </PixelButton>
      </form>

      <PixelModal
        open={privacyOpen}
        onClose={() => setPrivacyOpen(false)}
        title="Privacy and Personal Data Consent"
      >
        <p>
          By continuing, you acknowledge that Fisher Timer will collect and use your
          personal data, including your Google account email, display name, and
          account information, for the purposes of creating and managing your
          account, providing study session features, displaying your profile and
          statistics, managing rewards and leaderboard features, and maintaining the
          security of the service.
        </p>
        <p>
          Your personal data will be handled in accordance with Thailand&rsquo;s
          Personal Data Protection Act (PDPA) and our Privacy Policy.
        </p>
        <p>
          You may update certain profile information and exercise your applicable
          personal data rights through your account or by contacting the service
          administrator.
        </p>
      </PixelModal>
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
