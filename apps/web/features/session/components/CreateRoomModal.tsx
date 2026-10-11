"use client";

import { FormEvent, useState } from "react";
import { PixelAlert } from "../../../components/ui/PixelAlert";
import { PixelButton } from "../../../components/ui/PixelButton";
import { PixelInput } from "../../../components/ui/PixelInput";
import { PixelModal } from "../../../components/ui/PixelModal";
import { ROOM_RULES, validateRoom } from "../session.api";

const SEAT_CAPTIONS: Record<number, string> = { 1: "solo", 2: "pair", 3: "trio", 4: "crew", 5: "full dock" };

export interface CreateRoomModalProps {
  open: boolean;
  onClose: () => void;
  /** Resolves with an error message to show, or null on success. */
  onCreate: (name: string, limit: number) => Promise<string | null>;
}

/**
 * UC-01: name the room and pick how many anglers fit on the dock. The limit
 * defaults to the maximum; 1 makes a solo room nobody else can join.
 */
export function CreateRoomModal({ open, onClose, onCreate }: CreateRoomModalProps) {
  const [name, setName] = useState("");
  const [limit, setLimit] = useState<number>(ROOM_RULES.maxLimit);
  const [touched, setTouched] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  const errors = validateRoom(name, limit);
  const showErrors = touched ? errors : {};

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setTouched(true);
    if (errors.name || errors.limit) return;
    setSubmitting(true);
    setServerError(null);
    const error = await onCreate(name.trim(), limit);
    setSubmitting(false);
    if (error) setServerError(error);
    else {
      setName("");
      setLimit(ROOM_RULES.maxLimit);
      setTouched(false);
    }
  }

  return (
    <PixelModal open={open} onClose={onClose} title="Open a new room">
      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-6">
        <PixelInput
          label="Room name"
          placeholder="e.g. Midterm grind"
          value={name}
          maxLength={ROOM_RULES.maxNameLength + 10}
          onChange={(e) => setName(e.target.value)}
          onBlur={() => setTouched(true)}
          error={showErrors.name}
          hint={`${[...name.trim()].length}/${ROOM_RULES.maxNameLength} · every room is public`}
          autoFocus
        />

        <fieldset className="flex flex-col gap-2">
          <legend className="pixel-label">Spots on the dock</legend>
          <div className="pixel-seat-picker">
            {Array.from({ length: ROOM_RULES.maxLimit }, (_, i) => i + 1).map((n) => (
              <label key={n}>
                <input
                  type="radio"
                  name="participant_limit"
                  value={n}
                  checked={limit === n}
                  onChange={() => setLimit(n)}
                  className="pixel-seat-picker__input"
                />
                <span className="pixel-seat-picker__option">
                  <span className="pixel-seat-picker__count">{n}</span>
                  <span className="pixel-seat-picker__caption">{SEAT_CAPTIONS[n]}</span>
                </span>
              </label>
            ))}
          </div>
          <p className="text-sm text-bark">
            {limit === 1
              ? "Just you on the dock. Nobody else can join."
              : `You and up to ${limit - 1} other${limit - 1 > 1 ? "s" : ""}. More anglers, rarer fish.`}
          </p>
          {showErrors.limit && <p className="pixel-error">{showErrors.limit}</p>}
        </fieldset>

        {serverError && <PixelAlert>{serverError}</PixelAlert>}

        <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <PixelButton variant="ghost" onClick={onClose}>
            Cancel
          </PixelButton>
          <PixelButton type="submit" disabled={submitting}>
            {submitting ? "Casting off…" : "Open room"}
          </PixelButton>
        </div>
      </form>
    </PixelModal>
  );
}
