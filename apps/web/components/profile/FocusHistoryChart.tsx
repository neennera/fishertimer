"use client";

import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type CSSProperties,
  type KeyboardEvent,
} from "react";
import { createPortal } from "react-dom";
import { cx } from "../../lib/cx";
import { formatFocusMinutes } from "../../lib/format";
import type { DailyFocus } from "../../lib/profile-types";
import { PixelButton } from "../ui/PixelButton";

const RANGES = [7, 14, 30] as const;
type Range = (typeof RANGES)[number];
const LABEL_EVERY: Record<Range, number> = { 7: 1, 14: 2, 30: 5 };
const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

// Dates stay plain YYYY-MM-DD strings: new Date("YYYY-MM-DD") parses as UTC
// midnight and can land on the wrong day locally.
function dateParts(date: string) {
  const [y = 0, m = 1, d = 1] = date.split("-").map(Number);
  return { y, m, d };
}

function axisLabel(date: string) {
  const { m, d } = dateParts(date);
  return `${d}/${m}`;
}

function longLabel(date: string) {
  const { y, m, d } = dateParts(date);
  const weekday = new Date(y, m - 1, d).toLocaleDateString("en-GB", { weekday: "short" });
  return `${weekday} ${d} ${MONTHS[m - 1]}`;
}

// Clean ticks: the smallest step giving at most 4 ticks; the top is at
// least 60 so a nearly empty range doesn't blow up one bar.
function scale(max: number) {
  for (const step of [30, 60, 120, 180, 240, 360, 480, 720]) {
    const top = Math.max(60, Math.ceil(max / step) * step);
    if (top / step <= 4) {
      return { top, ticks: Array.from({ length: top / step + 1 }, (_, i) => i * step) };
    }
  }
  const top = Math.ceil(max / 240) * 240;
  return { top, ticks: [0, top / 4, top / 2, (top * 3) / 4, top] };
}

export function focusSummary(days: DailyFocus[], range: number) {
  const shown = days.slice(-range);
  const total = shown.reduce((sum, day) => sum + day.focus_minutes, 0);
  return { total, perDay: Math.round(total / range) };
}

/** days: the 30-day history, or null while loading. */
export function FocusHistoryChart({ days, titleId }: { days: DailyFocus[] | null; titleId: string }) {
  const [range, setRange] = useState<Range>(14);
  const [active, setActive] = useState<number | null>(null);
  const [focusIndex, setFocusIndex] = useState(29);
  const [grown, setGrown] = useState(false);
  const [bubble, setBubble] = useState<{ left: number; top: number } | null>(null);
  const plotRef = useRef<HTMLDivElement>(null);
  const areaRef = useRef<HTMLDivElement>(null);
  const bubbleRef = useRef<HTMLDivElement>(null);
  const colRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const loaded = days !== null;
  const all = days ?? [];
  const first = all.length - range;
  const shown = all.slice(-range);
  const { total, perDay } = focusSummary(all, range);
  const { top, ticks } = scale(Math.max(0, ...shown.map((d) => d.focus_minutes)));
  const empty = loaded && total === 0;

  useEffect(() => {
    if (!loaded) {
      return;
    }
    const id = requestAnimationFrame(() => setGrown(true));
    return () => cancelAnimationFrame(id);
  }, [loaded]);

  // Centred above its bar; may overhang the panel but stays inside the viewport.
  useLayoutEffect(() => {
    const plot = plotRef.current;
    const area = areaRef.current;
    const el = bubbleRef.current;
    const col = active === null ? null : colRefs.current[active];
    const day = active === null ? null : all[active];
    if (!plot || !area || !el || !col || !day) {
      setBubble(null);
      return;
    }
    const artPx = parseFloat(getComputedStyle(plot).getPropertyValue("--px")) || 1;
    const a = area.getBoundingClientRect();
    const c = col.getBoundingClientRect();
    const barTop = a.top + a.height * (1 - Math.min(1, day.focus_minutes / top));
    const w = el.offsetWidth;
    const edge = 2 * artPx;
    const left = Math.min(Math.max(edge, c.left + c.width / 2 - w / 2), window.innerWidth - edge - w);
    setBubble({
      left: left + window.scrollX,
      top: barTop - 2 * artPx - el.offsetHeight + window.scrollY,
    });
    // `all` changes identity every render; the active day and scale are what matter.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [active, range, top]);

  function moveFocus(from: number, step: number) {
    const next = Math.min(all.length - 1, Math.max(first, from + step));
    setFocusIndex(next);
    colRefs.current[next]?.focus();
  }

  function onKeyDown(event: KeyboardEvent, index: number) {
    if (event.key === "ArrowLeft" || event.key === "ArrowRight") {
      event.preventDefault();
      moveFocus(index, event.key === "ArrowLeft" ? -1 : 1);
    }
  }

  const activeDay = active === null ? null : all[active];

  return (
    <div className="font-label">
      <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
        <h2 id={titleId} className="font-display text-2xl leading-none">
          Focus history
        </h2>
        <div className="flex gap-2" role="group" aria-label="Range">
          {RANGES.map((r) => (
            <PixelButton
              key={r}
              small
              variant={r === range ? "primary" : "ghost"}
              aria-pressed={r === range}
              disabled={!loaded}
              onClick={() => {
                setRange(r);
                setActive(null);
                setFocusIndex(29);
              }}
            >
              {r}D
            </PixelButton>
          ))}
        </div>
      </div>
      <div className="mt-2 flex items-baseline justify-between gap-4 font-body text-sm text-bark">
        {loaded ? (
          <>
            <p>
              <span className="font-bold">Total</span> <span className="text-ink">{formatFocusMinutes(total)}</span>
            </p>
            <p>
              <span className="font-bold">Avg</span> <span className="text-ink">{formatFocusMinutes(perDay)}</span>/day
            </p>
          </>
        ) : (
          <>
            <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-24" />
            <span aria-hidden="true" className="pixel-skeleton pixel-skeleton--text w-24" />
          </>
        )}
      </div>

      <div className="pixel-chart mt-4">
        <div className="pixel-chart__y" aria-hidden="true">
          {ticks.map((tick) => (
            <span key={tick} className="pixel-chart__tick" style={{ "--at": tick / top } as CSSProperties}>
              {tick}m
            </span>
          ))}
        </div>

        <div
          ref={plotRef}
          className="pixel-chart__plot"
          role="group"
          aria-label={
            loaded
              ? `Focus history, last ${range} days: ${formatFocusMinutes(total)} total, ${formatFocusMinutes(perDay)} per day`
              : "Focus history, loading"
          }
          onMouseLeave={() => setActive(null)}
        >
          <div ref={areaRef} className="pixel-chart__area" aria-hidden="true">
            {ticks.map((tick) => (
              <span key={tick} className="pixel-chart__grid" style={{ "--at": tick / top } as CSSProperties} />
            ))}
            {empty && <p className="pixel-chart__empty">No focus time in the last {range} days</p>}
          </div>

          <div className="pixel-chart__cols">
            {all.map((day, index) => {
              const isShown = index >= first;
              const position = index - first;
              const labelled = isShown && (all.length - 1 - index) % LABEL_EVERY[range] === 0;
              return (
                <button
                  key={day.date}
                  ref={(el) => {
                    colRefs.current[index] = el;
                  }}
                  type="button"
                  className={cx(
                    "pixel-chart__col",
                    !isShown && "pixel-chart__col--hidden",
                    active === index && "pixel-chart__col--active",
                  )}
                  style={
                    {
                      "--h": grown ? Math.min(1, day.focus_minutes / top) : 0,
                      "--i": Math.max(0, position),
                    } as CSSProperties
                  }
                  tabIndex={isShown && index === focusIndex ? 0 : -1}
                  aria-hidden={isShown ? undefined : true}
                  aria-label={`${longLabel(day.date)}: ${day.focus_minutes} minutes`}
                  onMouseEnter={() => setActive(index)}
                  onFocus={() => {
                    setActive(index);
                    setFocusIndex(index);
                  }}
                  onBlur={() => setActive(null)}
                  onClick={() => setActive(index)}
                  onKeyDown={(event) => onKeyDown(event, index)}
                >
                  <span className="pixel-chart__bar-slot">
                    <span
                      className={cx("pixel-chart__bar", day.focus_minutes === 0 && "pixel-chart__bar--zero")}
                    />
                  </span>
                  {labelled && (
                    <span
                      className={cx(
                        "pixel-chart__label",
                        index === all.length - 1 && "pixel-chart__label--end",
                      )}
                    >
                      {axisLabel(day.date)}
                    </span>
                  )}
                </button>
              );
            })}
          </div>

          {activeDay &&
            createPortal(
              <div
                ref={bubbleRef}
                className="pixel-chart__bubble"
                aria-hidden="true"
                style={bubble ?? { visibility: "hidden" }}
              >
                <span className="pixel-chart__bubble-value">
                  {activeDay.focus_minutes === 0 ? "No focus" : formatFocusMinutes(activeDay.focus_minutes)}
                </span>
                <span className="pixel-chart__bubble-date">{longLabel(activeDay.date)}</span>
              </div>,
              document.body,
            )}
        </div>
      </div>

      {loaded && (
        <table className="sr-only">
          <caption>Focus minutes per day, last {range} days</caption>
          <thead>
            <tr>
              <th scope="col">Date</th>
              <th scope="col">Minutes</th>
            </tr>
          </thead>
          <tbody>
            {shown.map((day) => (
              <tr key={day.date}>
                <td>{longLabel(day.date)}</td>
                <td>{day.focus_minutes}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
