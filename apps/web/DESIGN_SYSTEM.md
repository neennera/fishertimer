# Fisher Timer Design System

How the UI is built in `apps/web`.

## Files

```
app/tokens.css     all colour and size values
app/pixel.css      the visual style, as CSS classes
components/ui/     React components that use those classes
app/**             pages
```

## Rules

1. Only `tokens.css` contains hex colours and px values. Everywhere else uses a
   token or `calc(var(--px) * n)`.
2. Pages and components do not write their own surface styling. Use a class
   from `pixel.css`, or add one there.
3. Tailwind for layout: flex, grid, gap, breakpoints. `pixel.css` for borders,
   fills and depth.
4. Use real `<input>`, `<button>` and `<label>`. No custom replacements.
5. Every clickable or focusable element needs a visible focus state.

Before opening a PR:

- [ ] No hex colours outside `tokens.css`
- [ ] No px values outside `tokens.css`
- [ ] Checked at 375px, 768px and 1280px
- [ ] Focus visible on everything interactive

## Size

```css
--px: 3px;
```

One pixel of the art style. Borders, corners, depth, focus rings, image scale
and padding are all multiples of it. Write `calc(var(--px) * n)`.

Changing `--px` resizes the whole interface.

Font sizes are the exception. Set them normally — web fonts do not fit the grid.

## Colour

Tokens are in the `@theme` block of `tokens.css`. Each one works as
`var(--color-name)` and as a Tailwind class such as `bg-amber` or `text-bark`.

**In use**

| Token | For |
| --- | --- |
| `--color-ink` | Text |
| `--color-bark` | Secondary text, edges |
| `--color-cream` | Panel background |
| `--color-cream-2` | Secondary buttons, dividers |
| `--color-amber` | Primary buttons |
| `--color-amber-dk` | Button hover and depth |
| `--color-stone` / `-dk` / `-ink` | Disabled button cap, depth, label |
| `--color-oak` | Header bar, wood |
| `--color-lake` | Accent, input focus |
| `--color-lake-dp` | Button focus ring |
| `--color-rust` | Errors, danger buttons |
| `--color-rust-dk` | Danger button hover and depth |
| `--color-rust-dp` | Danger button hover depth |

**Not used yet** — for background scenes, not for UI.

| Token | For |
| --- | --- |
| `--color-oak-lt` | Wood grain, floor |
| `--color-sage` | Grass |
| `--color-forest` | Trees |
| `--color-sky` | Sky, walls |

**Page colours** — `--color-bg`, `--color-fg`, `--color-muted`, `--color-rule`,
`--color-surface`, `--color-surface-2`.

These six are the only ones that change in dark mode. Panels stay cream and
buttons stay amber in both. Dark mode follows the operating system setting;
`data-theme="light"` or `data-theme="dark"` on `<html>` overrides it.

To add a colour, add a token. Do not put a hex value in a component.

## Type

| Token | Font | For |
| --- | --- | --- |
| `--font-display` | Jersey 15 | Titles and headings, 18px and up |
| `--font-numeric` | Jersey 25 | Numbers and button labels |
| `--font-label` | Silkscreen | Small uppercase labels, 9–12px |
| `--font-body` | DotGothic16 | Sentences |

Tailwind classes: `font-display`, `font-numeric`, `font-label`, `font-body`.
Use the token, not the font name.

All four are pixel fonts. DotGothic16 needs `line-height: 1.75`, set in
`globals.css`. Do not lower it.

To change a font, edit `fonts.ts` and the matching token in `tokens.css`.
Components do not change.

## Classes

From `app/pixel.css`. Use these before writing new CSS.

| Class | What it is |
| --- | --- |
| `.pixel-header` | Top bar. Fixed height, so it never shifts between pages. Parts: `__brand`, `__logo`, `__title`, `__avatar` |
| `.pixel-panel` | Panel background with an edge and a bottom band |
| `.pixel-btn` | Button. Add `--ghost`, `--danger` (red, destructive actions like sign out), `--icon` (square, one 24-unit pixelarticons glyph) or `--block` |
| `.pixel-label` | Uppercase field label |
| `.pixel-input` | Text field. Set `aria-invalid="true"` for the error state |
| `.pixel-error` | Message under a field |
| `.pixel-field-error` | Same slot as `.pixel-label` (font, tracking, case) but red. Always mounted with a reserved line-height, even with no message, so an error appearing doesn't shift the layout |
| `.pixel-checkbox` | Real `<input type="checkbox">`, visually hidden, plus a `__box` and `__label`. Checked state is fully filled amber, driven off the input's `:checked` via the adjacent-sibling selector |
| `.pixel-link` | A `<button>` styled to read as an inline text link (e.g. opening a modal from inside a sentence) |
| `.pixel-modal` | `.pixel-panel` surface for a dialog. Pair with `.pixel-modal-overlay` (fixed, dimmed backdrop), `__title`, `__close` (red on hover), `__body` |
| `.pixel-alert` | Status message. Add `--warn` |
| `.pixel-badge` | Small tag |
| `.pixel-tile` | Number and caption, with `__value` and `__caption` |
| `.pixel-sprite` | Pixel image. Sets `image-rendering: pixelated` |

React components in `components/ui/`: `PixelPanel`, `PixelButton`,
`PixelInput`, `PixelCheckbox`, `PixelModal`, `PixelAlert`, `PixelBadge`,
`StatTile`.

**Modals** portal to `document.body` (`react-dom`'s `createPortal`) rather
than rendering in place — `position: fixed` is still clipped to the paint
region of any ancestor with its own `clip-path`, which every `.pixel-panel`
has, so an in-place overlay would be boxed into the panel instead of covering
the viewport.

**The style.** Square corners with one pixel cut off each corner
(`--pixclip`), a one-pixel inset edge, and a one-pixel darker band along the
bottom. No rounded corners and no blurred shadows.

**The button.** `::after` is the thick base, `::before` is the top face one
pixel above it. On press the top face moves down and the base is hidden. The
element keeps the same height, so nothing else moves.

To add a surface, add a class here using the same tokens.

## Layout

Tailwind. Mobile first. Breakpoints: `md` 768px, `lg` 1024px.

- Single-task pages: one centred panel at all widths.
- Grids of tiles or cards: one column on phones.
- Side-by-side panels: stacked on phones.

## Motion

The system has no animation of its own. What moves is decided per page.

Anything that animates must:

1. Look correct when stopped. Nothing should be invisible until it animates.
2. Work with `prefers-reduced-motion`. `globals.css` already turns off
   animations and transitions under that setting.

## Images

No images are used yet. Surfaces are drawn in CSS, so pages work without art.

When art is added:

- Save files at their real size. Do not enlarge them in the image editor.
- Display them at a whole multiple of a screen pixel, with
  `image-rendering: pixelated`.
- Keep text on a `.pixel-panel`, not directly on a background image.
- List the source in `CREDITS.md`. The repo is public, so assets that forbid
  redistribution cannot be committed.

```
public/sprites/ui/      panel and button art
public/sprites/scene/   background scenes
public/sprites/items/   small images
```

**Replacing CSS surfaces with images.** Edit `pixel.css` only. No component or
page changes.

```css
.pixel-panel {
  border: calc(var(--px) * 6) solid transparent;
  border-image: url(/sprites/ui/panel.png) 6 fill / calc(var(--px) * 6) round;
  image-rendering: pixelated;
  /* replaces background, clip-path and box-shadow */
}
```

- `6` is the corner size in the image's own pixels. It must match the art.
- `calc(var(--px) * 6)` is that same 6, scaled up. The two always match.
- `fill` draws the middle of the image. Without it the panel is see-through.
- Use `round`, not `stretch`. `stretch` blurs the edges.

Button states swap the image:

```css
.pixel-btn:hover  { border-image-source: url(/sprites/ui/btn-hover.png); }
.pixel-btn:active { border-image-source: url(/sprites/ui/btn-down.png); }
```
