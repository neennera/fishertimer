import { cx } from "../../../lib/cx";
import { PixelButton } from "../../../components/ui/PixelButton";
import type { StudyRoom } from "../session.api";

export interface RoomCardProps {
  room: StudyRoom;
  /** The user is already in this room. */
  mine: boolean;
  busy: boolean;
  onJoin: () => void;
  onOpen: () => void;
}

/** One room on the list (UC-02 step 2): name, seats taken, and Join. */
export function RoomCard({ room, mine, busy, onJoin, onOpen }: RoomCardProps) {
  const full = room.participant_count >= room.participant_limit;
  const solo = room.participant_limit === 1;

  return (
    <li
      className={cx(
        "pixel-tile pixel-room-card",
        full && !mine && "pixel-room-card--full",
        mine && "pixel-room-card--mine",
      )}
    >
      <div className="min-w-0 flex-1">
        <p className="pixel-room-card__name" title={room.name}>
          {room.name}
        </p>
        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1">
          <span
            className="pixel-seats"
            role="img"
            aria-label={`${room.participant_count} of ${room.participant_limit} spots taken`}
          >
            {Array.from({ length: room.participant_limit }, (_, i) => (
              <span
                key={i}
                className={cx("pixel-seats__seat", i < room.participant_count && "pixel-seats__seat--taken")}
              />
            ))}
          </span>
          <span className="pixel-room-card__meta">
            {room.participant_count}/{room.participant_limit} anglers
            {solo && " · solo"}
          </span>
        </div>
      </div>

      {mine ? (
        <PixelButton onClick={onOpen}>Return</PixelButton>
      ) : (
        <PixelButton onClick={onJoin} disabled={full || busy} aria-label={`Join ${room.name}`}>
          {full ? "Full" : busy ? "Joining…" : "Join"}
        </PixelButton>
      )}
    </li>
  );
}
