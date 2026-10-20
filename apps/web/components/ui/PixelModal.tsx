"use client";

import { ReactNode, useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";
import { cx } from "../../lib/cx";

export interface PixelModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
}

/**
 * A pixel-panel surface over a dimmed backdrop. Closes on Escape or a
 * backdrop click; focuses its close button on open so keyboard/screen-reader
 * users land inside the dialog rather than on whatever triggered it.
 */
export function PixelModal({ open, onClose, title, children }: PixelModalProps) {
  const closeRef = useRef<HTMLButtonElement>(null);
  const dialogRef = useRef<HTMLDivElement>(null);
  const titleId = useId();
  // Callers usually pass an inline arrow, so onClose changes on every parent
  // render. Reading it through a ref keeps the effect below tied to `open`
  // only; otherwise a re-render (e.g. a polling list behind the dialog)
  // would re-focus the close button mid-typing, and the next space or Enter
  // would close the dialog.
  const onCloseRef = useRef(onClose);
  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    if (!open) {
      return;
    }

    // Whatever had focus (the button that opened the dialog) gets it back on close.
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    closeRef.current?.focus();

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onCloseRef.current();
        return;
      }
      if (event.key !== "Tab" || !dialogRef.current) {
        return;
      }
      // Keep Tab inside the dialog: wrap from the last control to the first.
      const controls = Array.from(
        dialogRef.current.querySelectorAll<HTMLElement>(
          'button:not([disabled]), input:not([disabled]), a[href], select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ),
      );
      const first = controls[0];
      const last = controls[controls.length - 1];
      if (!first || !last) {
        return;
      }
      const active = document.activeElement;
      if (event.shiftKey && (active === first || !dialogRef.current.contains(active))) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && (active === last || !dialogRef.current.contains(active))) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      if (opener?.isConnected) {
        opener.focus();
      }
    };
  }, [open]);

  if (!open) {
    return null;
  }

  // Portalled to <body> rather than rendered in place: a `position: fixed`
  // overlay is still clipped to the paint region of any ancestor with its
  // own clip-path (every .pixel-panel has one) — a portal is the only way
  // out of that so the backdrop can cover the whole viewport.
  return createPortal(
    <div
      className="pixel-modal-overlay"
      onClick={(event) => {
        if (event.target === event.currentTarget) {
          onClose();
        }
      }}
    >
      <div
        ref={dialogRef}
        className={cx("pixel-panel", "pixel-modal")}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
      >
        <div className="flex items-start justify-between gap-4">
          <h2 id={titleId} className="pixel-modal__title">
            {title}
          </h2>
          <button
            ref={closeRef}
            type="button"
            className="pixel-modal__close h-11 w-11 sm:h-[calc(var(--px)*8)] sm:w-[calc(var(--px)*8)]"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </div>
        <div className="pixel-modal__body">{children}</div>
      </div>
    </div>,
    document.body,
  );
}
