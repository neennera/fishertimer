"use client";

import { useEffect, useId, useRef, useState, type FormEvent, type KeyboardEvent } from "react";
import { Check } from "pixelarticons/react/Check";
import { Close } from "pixelarticons/react/Close";
import { Loader } from "pixelarticons/react/Loader";
import { Pencil } from "pixelarticons/react/Pencil";
import { cx } from "../../lib/cx";
import { PixelButton } from "../ui/PixelButton";
import { updateDisplayName, type SessionUser } from "../../lib/auth";
import { validateDisplayName } from "../../lib/validate-display-name";

export interface DisplayNameProps {
  name: string;
  email: string;
  onSaved: (user: SessionUser) => void;
}

// Same type classes in both modes, so the text doesn't move when it becomes
// editable.
const NAME_TYPE = "font-display text-3xl leading-none";

/**
 * The display name with in-place editing (UC-07, 04a–04f states): the pencil
 * turns the name itself into an input, with check / close in the pencil's
 * place. Errors show between the name and the e-mail.
 */
export function DisplayName({ name, email, onSaved }: DisplayNameProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(name);
  const [attemptedSubmit, setAttemptedSubmit] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const pencilRef = useRef<HTMLButtonElement>(null);
  const returnFocus = useRef(false);
  const errorId = useId();

  useEffect(() => {
    if (editing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    } else if (returnFocus.current) {
      returnFocus.current = false;
      pencilRef.current?.focus();
    }
  }, [editing]);

  // Same rule as first-time setup: "too long" shows live, "empty" only once
  // they've tried to save.
  const liveError = validateDisplayName(draft);
  const shownError =
    (attemptedSubmit || draft.trim().length > 0 ? liveError : undefined) ?? saveError;

  function startEditing() {
    setDraft(name);
    setAttemptedSubmit(false);
    setSaveError(null);
    setEditing(true);
  }

  function stopEditing() {
    returnFocus.current = true;
    setEditing(false);
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setAttemptedSubmit(true);
    setSaveError(null);
    if (liveError) {
      return;
    }
    if (draft.trim() === name) {
      stopEditing();
      return;
    }

    setSaving(true);
    try {
      const result = await updateDisplayName(draft);
      if (result.ok) {
        onSaved(result.user);
        stopEditing();
      } else {
        setSaveError(result.error);
      }
    } catch {
      setSaveError("Couldn't save your changes. Please try again.");
    } finally {
      setSaving(false);
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape" && !saving) {
      stopEditing();
    }
  }

  return (
    <>
      {editing ? (
        <form className="pixel-name-row" onSubmit={handleSubmit} onKeyDown={handleKeyDown} noValidate>
          <span className={cx("pixel-inline-field", shownError && "pixel-inline-field--invalid")}>
            <input
              ref={inputRef}
              className={cx("pixel-inline-input", NAME_TYPE)}
              value={draft}
              onChange={(event) => {
                setDraft(event.target.value);
                setSaveError(null);
              }}
              disabled={saving}
              aria-label="Display name"
              aria-invalid={shownError ? "true" : undefined}
              aria-describedby={shownError ? errorId : undefined}
              maxLength={100}
              // Width fallback where field-sizing isn't supported.
              size={Math.max(8, draft.length + 1)}
            />
          </span>
          <PixelButton
            type="submit"
            variant="ghost"
            icon
            small
            className={cx(saving && "pixel-btn--busy")}
            disabled={saving}
            aria-label="Save display name"
            title="Save"
          >
            {saving ? <Loader aria-hidden="true" /> : <Check aria-hidden="true" />}
          </PixelButton>
          <PixelButton
            variant="ghost"
            icon
            small
            onClick={stopEditing}
            disabled={saving}
            aria-label="Cancel editing"
            title="Cancel"
          >
            <Close aria-hidden="true" />
          </PixelButton>
        </form>
      ) : (
        <div className="pixel-name-row">
          <p className={cx("pixel-name-row__text", NAME_TYPE)} title={name}>
            {name}
          </p>
          <PixelButton
            ref={pencilRef}
            variant="ghost"
            icon
            small
            onClick={startEditing}
            aria-label="Edit display name"
            title="Edit display name"
          >
            <Pencil aria-hidden="true" />
          </PixelButton>
        </div>
      )}
      <p id={errorId} className="pixel-inline-error" role={editing && shownError ? "alert" : undefined}>
        {editing ? shownError : null}
      </p>
      <p className="text-sm text-bark break-words">{email}</p>
    </>
  );
}
