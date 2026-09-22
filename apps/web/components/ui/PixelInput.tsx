"use client";

import { InputHTMLAttributes, useId } from "react";
import { cx } from "../../lib/cx";

export interface PixelInputProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "id"> {
  label: string;
  /** Validation message. Its presence is what puts the field in the error state. */
  error?: string;
  /** Help text under the field, e.g. "from Google, not editable". */
  hint?: string;
  id?: string;
}

/**
 * A native input with a pixel surface. Deliberately native: tab order,
 * autofill, password managers and screen readers all keep working.
 */
export function PixelInput({
  label,
  error,
  hint,
  id,
  className,
  ...props
}: PixelInputProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;
  const errorId = `${inputId}-error`;
  const hintId = `${inputId}-hint`;

  const describedBy =
    [error ? errorId : null, hint ? hintId : null].filter(Boolean).join(" ") ||
    undefined;

  return (
    <div className={cx("w-full", className)}>
      <label className="pixel-label" htmlFor={inputId}>
        {label}
      </label>
      <input
        id={inputId}
        className="pixel-input"
        aria-invalid={error ? "true" : undefined}
        aria-describedby={describedBy}
        {...props}
      />
      {hint && !error && (
        <p id={hintId} className="mt-2 text-sm text-bark">
          {hint}
        </p>
      )}
      {error && (
        <p id={errorId} className="pixel-error" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
