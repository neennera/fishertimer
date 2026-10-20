import { HTMLAttributes } from "react";
import { cx } from "../../lib/cx";

export interface PixelAlertProps extends HTMLAttributes<HTMLDivElement> {
  /** `error` for a failed action, `warn` for something recoverable. */
  tone?: "error" | "warn";
}

/**
 * Carries the exceptional flows from UC-06 and UC-07: consent denied, code
 * exchange failed, account creation failed, save failed.
 *
 * Write the message as what happened and what to do — never an apology.
 * An `error` is announced to screen readers straight away (role="alert"), a
 * `warn` politely (role="status"); pass `role` to override.
 */
export function PixelAlert({
  tone = "error",
  className,
  ...props
}: PixelAlertProps) {
  return (
    <div
      role={tone === "error" ? "alert" : "status"}
      className={cx(
        "pixel-alert",
        tone === "warn" && "pixel-alert--warn",
        className,
      )}
      {...props}
    />
  );
}
