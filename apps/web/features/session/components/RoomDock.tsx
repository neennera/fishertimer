import type { CSSProperties } from "react";
import { cx } from "../../../lib/cx";
import { initials } from "../../../lib/initials";
import type { TimerState } from "../../timer/timer.api";
import type { RoomParticipant } from "../session.api";
import { ANGLER_COPY, anglerColor, anglerState } from "./angler";

export interface RoomDockProps {
  participants: RoomParticipant[];
  limit: number;
  timers: Record<string, TimerState>;
  meId: string;
}

/**
 * The room as a pier on the lake: one spot per seat, each participant an
 * angler whose line and bobber follow their timer (pixel.css .pixel-dock).
 * The picture is decoration; the roster under it carries the same facts
 * for screen readers.
 */
export function RoomDock({ participants, limit, timers, meId }: RoomDockProps) {
  const spots = Array.from({ length: Math.max(limit, participants.length) }, (_, i) => participants[i] ?? null);

  return (
    <div className="flex flex-1 flex-col gap-4">
      <div className="pixel-dock flex-1" aria-hidden="true">
        <div className="pixel-dock__scene">
          <div className="pixel-dock__spots" style={{ "--spots": spots.length } as CSSProperties}>
            {spots.map((p, i) => {
              if (!p) {
                return (
                  <div key={`open-${i}`} className="pixel-angler pixel-angler--open" data-state="idle">
                    <span className="pixel-angler__tag">open</span>
                    <span className="pixel-angler__head" />
                    <span className="pixel-angler__body" />
                  </div>
                );
              }
              const state = anglerState(timers[p.user_id]);
              return (
                <div
                  key={p.user_id}
                  className={cx("pixel-angler", p.user_id === meId && "pixel-angler--me")}
                  data-state={state}
                  style={{ "--angler": anglerColor(p.user_id) } as CSSProperties}
                >
                  <span className="pixel-angler__tag">{p.user_id === meId ? "You" : p.display_name || "Angler"}</span>
                  <span className="pixel-angler__mark">{ANGLER_COPY[state].mark}</span>
                  <span className="pixel-angler__hat" />
                  <span className="pixel-angler__head" />
                  <span className="pixel-angler__body">{initials(p.display_name || "?")}</span>
                  <span className="pixel-angler__rod" />
                  <span className="pixel-angler__line" />
                  <span className="pixel-angler__ripple" />
                  <span className="pixel-angler__bobber" />
                </div>
              );
            })}
          </div>
        </div>
      </div>

      <ul className="pixel-roster" aria-label="Anglers in this room">
        {participants.map((p) => {
          const state = anglerState(timers[p.user_id]);
          const copy = ANGLER_COPY[state];
          return (
            <li key={p.user_id} className="pixel-roster__row">
              <span className="pixel-roster__swatch" style={{ background: anglerColor(p.user_id) }} aria-hidden="true">
                {initials(p.display_name || "?")}
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate font-display text-xl leading-tight">
                  {p.display_name || "Angler"}
                  {p.user_id === meId && <span className="ml-2 font-label text-xs text-bark">(you)</span>}
                </span>
                <span className="block font-label text-[10px] uppercase tracking-wider text-bark">
                  since {formatTime(p.joined_at)}
                </span>
              </span>
              <span className={cx("pixel-chip", copy.chipClass)}>{copy.chip}</span>
            </li>
          );
        })}
      </ul>
    </div>
  );
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? "—" : d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}
