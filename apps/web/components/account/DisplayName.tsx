"use client";

import {
  useEffect,
  useId,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent,
  type ReactNode,
} from "react";
import { Check } from "pixelarticons/react/Check";
import { Close } from "pixelarticons/react/Close";
import { Loader } from "pixelarticons/react/Loader";
import { Pencil } from "pixelarticons/react/Pencil";
import { cx } from "../../lib/cx";
import { PixelButton } from "../ui/PixelButton";
import { updateDisplayName, type SessionUser } from "../../lib/auth";
import { validateDisplayName } from "../../lib/validate-display-name";

export type DisplayNameProps =
  | {
      name: string;
      editable?: false;
      /** Adds the bottom line (as in edit mode) with this at its end, so the
       * name and the action sit where they do on /account. */
      aside?: ReactNode;
    }
  | {
      name: string;
      editable: true;
      email: string;
      onSaved: (user: SessionUser) => void;
      /** The session ended while saving (401). */
      onSignedOut: () => void;
      /** Shown at the end of the e-mail line. */
      aside?: ReactNode;
    };

// The line under the name: the e-mail (or other detail) with an action at
// its end. The reserved error line above it keeps both modes the same height.
function BottomLine({ detail, aside }: { detail?: ReactNode; aside?: ReactNode }) {
  return (
    <div className="pixel-name-bottom flex flex-wrap items-center justify-between gap-x-4 gap-y-1 text-sm">
      <p className="min-w-0 text-bark break-words">{detail}</p>
      {aside}
    </div>
  );
}

// Same type classes in both modes, so the text doesn't move when it becomes
// editable.
const NAME_TYPE = "font-display text-3xl leading-none";

/**
 * The display name, read-only or (for the owner) with in-place editing
 * (UC-07, 04a–04f states): the pencil turns the name itself into an input,
 * with check / close in the pencil's place. Errors show between the name and
 * the e-mail.
 */
export function DisplayName(props: DisplayNameProps) {
  const { name } = props;
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
        if (props.editable) {
          props.onSaved(result.user);
        }
        stopEditing();
      } else if (result.code === "signed_out") {
        if (props.editable) {
          props.onSignedOut();
        }
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

  const nameText = (
    <p className={cx("pixel-name-row__text", NAME_TYPE)} title={name}>
      {name}
    </p>
  );

  if (!props.editable) {
    return (
      <>
        <div className="pixel-name-row">{nameText}</div>
        {props.aside && (
          <>
            <p className="pixel-inline-error" aria-hidden="true" />
            <BottomLine aside={props.aside} />
          </>
        )}
      </>
    );
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
          {nameText}
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
      <BottomLine detail={props.email} aside={props.aside} />
    </>
  );
}
