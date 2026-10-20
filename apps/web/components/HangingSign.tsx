export interface HangingSignProps {
  /** Text on the board. */
  text?: string;
}

/**
 * The wooden "Fisher Timer" signboard on its post, with the bird perched at
 * the right. Place it inside a `relative` wrapper: it sits `bottom-full`, so it
 * stands on the top edge of whatever panel the wrapper holds.
 *
 * Decorative (the page has its own heading), so it is hidden from screen readers.
 */
export function HangingSign({ text = "Fisher Timer" }: HangingSignProps) {
  return (
    <div aria-hidden="true" className="absolute inset-x-0 bottom-full flex flex-col">
      <div className="pixel-sign__board mx-auto">
        <span className="pixel-sign__board-text">{text}</span>
      </div>
      <div className="pixel-sign__post mx-auto" />
      <div className="pixel-sign-bird self-end" />
    </div>
  );
}
