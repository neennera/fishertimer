"use client";

import { Minus } from "pixelarticons/react/Minus";
import { Plus } from "pixelarticons/react/Plus";
import { PixelButton } from "../../components/ui/PixelButton";
import { clampMinutes } from "./timer-model";

export interface DurationPickerProps {
  /** Used for the radio group name and labels. */
  label: string;
  value: number;
  onChange: (minutes: number) => void;
  presets: number[];
  min: number;
  max: number;
  /** Step for the - / + keys. */
  step?: number;
}

/**
 * Picks a cycle length (UC-05 step 2 / 10): a row of presets plus - / +
 * for anything in between, always kept inside the server's allowed range
 * (E-1). Presets are real radio inputs (.pixel-seat-picker).
 */
export function DurationPicker({ label, value, onChange, presets, min, max, step = 5 }: DurationPickerProps) {
  const options = presets.filter((p) => p >= min && p <= max);
  const name = `duration-${label.toLowerCase().replace(/\s+/g, "-")}`;
  const down = clampMinutes(value > step ? value - step : value - 1, min, max);
  const up = clampMinutes(value < step ? step : value + step, min, max);

  return (
    <fieldset className="flex flex-col gap-3">
      <legend className="pixel-label">{label}</legend>
      <div className="pixel-seat-picker" style={{ gridTemplateColumns: `repeat(${options.length}, minmax(0, 1fr))` }}>
        {options.map((minutes) => (
          <label key={minutes}>
            <input
              type="radio"
              name={name}
              value={minutes}
              checked={value === minutes}
              onChange={() => onChange(minutes)}
              className="pixel-seat-picker__input"
            />
            <span className="pixel-seat-picker__option">
              <span className="pixel-seat-picker__count">{minutes}</span>
              <span className="pixel-seat-picker__caption">min</span>
            </span>
          </label>
        ))}
      </div>
      <div className="flex items-center justify-center gap-3">
        <PixelButton
          variant="ghost"
          icon
          small
          aria-label={`${label}: ${down} minutes`}
          disabled={value <= min}
          onClick={() => onChange(down)}
        >
          <Minus aria-hidden="true" />
        </PixelButton>
        <output className="pixel-stepper__value" aria-live="polite">
          {value} <span className="pixel-stepper__unit">min</span>
        </output>
        <PixelButton
          variant="ghost"
          icon
          small
          aria-label={`${label}: ${up} minutes`}
          disabled={value >= max}
          onClick={() => onChange(up)}
        >
          <Plus aria-hidden="true" />
        </PixelButton>
      </div>
    </fieldset>
  );
}
