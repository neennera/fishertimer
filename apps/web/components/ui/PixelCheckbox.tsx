"use client";

import { InputHTMLAttributes, ReactNode, useId } from "react";
import { cx } from "../../lib/cx";

export interface PixelCheckboxProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "id" | "type"> {
  id?: string;
  /** Label content — can include inline links (e.g. to open a modal). */
  children: ReactNode;
}

/**
 * A real `<input type="checkbox">`, visually hidden but still tabbable and
 * screen-reader operable, paired with a pixel-styled box drawn from its
 * `:checked` state via the adjacent-sibling selector in pixel.css.
 */
export function PixelCheckbox({
  id,
  className,
  children,
  ...props
}: PixelCheckboxProps) {
  const generatedId = useId();
  const inputId = id ?? generatedId;

  return (
    <div className={cx("pixel-checkbox", className)}>
      <input id={inputId} type="checkbox" className="pixel-checkbox__input" {...props} />
      <label htmlFor={inputId} className="pixel-checkbox__box" aria-hidden="true" />
      <label htmlFor={inputId} className="pixel-checkbox__label">
        {children}
      </label>
    </div>
  );
}
