"use client";

import { ReactNode, useEffect, useRef } from "react";
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

  useEffect(() => {
    if (!open) {
      return;
    }

    closeRef.current?.focus();

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

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
        className={cx("pixel-panel", "pixel-modal")}
        role="dialog"
        aria-modal="true"
        aria-labelledby="pixel-modal-title"
      >
        <div className="flex items-start justify-between gap-4">
          <h2 id="pixel-modal-title" className="pixel-modal__title">
            {title}
          </h2>
          <button
            ref={closeRef}
            type="button"
            className="pixel-modal__close"
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
