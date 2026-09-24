import { ButtonHTMLAttributes } from "react";
import { cx } from "../../lib/cx";

type Variant = "primary" | "ghost";

export interface PixelButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  /** Fill the width of its container — used for the sign-in and Continue buttons. */
  block?: boolean;
}

/**
 * A chunky key sitting in its own socket. Pressing drops the cap by one art
 * pixel and the thickness vanishes with it, so the key ends up flush.
 *
 * Always pass an explicit `type` — an unlabelled button inside a form submits it.
 */
export function PixelButton({
  variant = "primary",
  block = false,
  className,
  type = "button",
  ...props
}: PixelButtonProps) {
  return (
    <button
      type={type}
      className={cx(
        "pixel-btn",
        variant === "ghost" && "pixel-btn--ghost",
        block && "pixel-btn--block",
        className,
      )}
      {...props}
    />
  );
}
