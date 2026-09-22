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
 */
export function PixelAlert({
  tone = "error",
  className,
  ...props
}: PixelAlertProps) {
  return (
    <div
      role="status"
      className={cx(
        "pixel-alert",
        tone === "warn" && "pixel-alert--warn",
        className,
      )}
      {...props}
    />
  );
}
