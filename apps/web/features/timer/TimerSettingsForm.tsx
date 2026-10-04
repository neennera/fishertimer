"use client";

import { FormEvent, useState } from "react";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelButton } from "../../components/ui/PixelButton";
import { DurationPicker } from "./DurationPicker";
import { saveTimerSettings, type TimerState } from "./timer.api";

export const WORK_PRESETS = [15, 25, 45, 60];
export const REST_PRESETS = [5, 10, 15];

export interface TimerSettingsFormProps {
  /** Current settings and allowed ranges (any TimerState of the user). */
  timer: TimerState;
  userId: string;
  /** The room, if the form is used inside one; settings are per user. */
  sessionId?: string;
  onSaved?: (state: TimerState) => void;
  onCancel?: () => void;
}

/** TimerSetting (UpdateTimerSetting): the default work and break lengths. */
export function TimerSettingsForm({ timer, userId, sessionId = "", onSaved, onCancel }: TimerSettingsFormProps) {
  const [work, setWork] = useState(timer.work_minutes);
  const [rest, setRest] = useState(timer.rest_minutes);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const changed = work !== timer.work_minutes || rest !== timer.rest_minutes;

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const state = await saveTimerSettings({ sessionId, userId }, work, rest);
      setSaved(true);
      onSaved?.(state);
    } catch {
      setError("Couldn't save your settings. Try again.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <DurationPicker
        label="Focus block"
        value={work}
        onChange={(m) => {
          setWork(m);
          setSaved(false);
        }}
        presets={WORK_PRESETS}
        min={timer.min_work_minutes}
        max={timer.max_work_minutes}
      />
      <DurationPicker
        label="Break"
        value={rest}
        onChange={(m) => {
          setRest(m);
          setSaved(false);
        }}
        presets={REST_PRESETS}
        min={timer.min_rest_minutes}
        max={timer.max_rest_minutes}
      />
      <p className="text-sm text-bark">
        These are your starting lengths. You can still pick a different length before each block. Focus blocks of 15
        minutes or more catch fish.
      </p>
      {error && <PixelAlert>{error}</PixelAlert>}
      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-end">
        {saved && !changed && <span className="font-label text-[10px] uppercase tracking-wider text-forest">Saved</span>}
        {onCancel && (
          <PixelButton variant="ghost" onClick={onCancel}>
            Close
          </PixelButton>
        )}
        <PixelButton type="submit" disabled={saving || !changed}>
          {saving ? "Saving…" : "Save defaults"}
        </PixelButton>
      </div>
    </form>
  );
}
