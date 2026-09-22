"use client";

import { useCallback, useEffect, useState } from "react";
import { Header } from "../../components/Header";
import { PixelAlert } from "../../components/ui/PixelAlert";
import { PixelBadge } from "../../components/ui/PixelBadge";
import { PixelButton } from "../../components/ui/PixelButton";
import { PixelInput } from "../../components/ui/PixelInput";
import { PixelPanel } from "../../components/ui/PixelPanel";
import { ParallaxScene } from "../../components/ui/ParallaxScene";
import { StatTile } from "../../components/ui/StatTile";
import { validateDisplayName } from "../../lib/validate-display-name";

const LAKE_LAYERS = [
  { src: "/sprites/scene/parallax-lake/sky.png", speed: 0.05 },
  { src: "/sprites/scene/parallax-lake/clouds.png", speed: 0.1 },
  { src: "/sprites/scene/parallax-lake/mountains.png", speed: 0.2 },
  { src: "/sprites/scene/parallax-lake/forest-far.png", speed: 0.35 },
  { src: "/sprites/scene/parallax-lake/forest-mid.png", speed: 0.5 },
  { src: "/sprites/scene/parallax-lake/forest-near.png", speed: 0.7 },
  { src: "/sprites/scene/parallax-lake/valley-fill.png", speed: 0.85 },
  { src: "/sprites/scene/parallax-lake/foreground.png", speed: 1 },
  { src: "/sprites/scene/parallax-lake/water.png", speed: 1.2 },
];

const ASSIGNED = [
  ["ink", "Text"],
  ["bark", "Secondary text, edges"],
  ["cream", "Panel"],
  ["cream-2", "Secondary control"],
  ["amber", "Primary action"],
  ["amber-dk", "Hover, depth"],
  ["oak", "Header, wood"],
  ["lake", "Accent, field focus"],
  ["lake-dp", "Focus ring"],
  ["rust", "Error"],
] as const;

const AVAILABLE = [
  ["oak-lt", "Wood grain"],
  ["sage", "Grass"],
  ["forest", "Trees"],
  ["sky", "Sky, walls"],
] as const;

const CHROME = [
  ["bg", "Page"],
  ["fg", "Page text"],
  ["muted", "Muted text"],
  ["rule", "Dividers"],
  ["surface", "Raised"],
  ["surface-2", "Sunken"],
] as const;

const TYPE_ROLES = [
  ["Display", "Jersey 15", "font-display", "text-3xl"],
  ["Figures", "Jersey 25", "font-numeric", "text-3xl"],
  ["Labels", "Silkscreen", "font-label", "text-xs tracking-widest"],
  ["Body", "DotGothic16", "font-body", "text-base"],
] as const;

const ALL_TOKENS = [...ASSIGNED, ...AVAILABLE, ...CHROME].map(([t]) => t);

type Theme = "light" | "dark" | "system";
const THEMES: readonly Theme[] = ["light", "dark", "system"];
const STORE_KEY = "fishertimer:styleguide-theme";

/**
 * Drives the same three states tokens.css supports: an explicit data-theme,
 * or no attribute at all, which falls through to prefers-color-scheme.
 */
function useTheme() {
  const [theme, setTheme] = useState<Theme>("system");

  useEffect(() => {
    try {
      const saved = localStorage.getItem(STORE_KEY);
      if (saved && (THEMES as readonly string[]).includes(saved)) {
        setTheme(saved as Theme);
      }
    } catch {
      /* private mode or blocked storage — system is a fine default */
    }
  }, []);

  useEffect(() => {
    const root = document.documentElement;
    if (theme === "system") root.removeAttribute("data-theme");
    else root.setAttribute("data-theme", theme);
    try {
      localStorage.setItem(STORE_KEY, theme);
    } catch {
      /* ignore */
    }
  }, [theme]);

  return [theme, setTheme] as const;
}

/** Reads live token values, so the page can never show a stale hex. */
function useTokenHex(theme: Theme) {
  const [hex, setHex] = useState<Record<string, string>>({});

  const read = useCallback(() => {
    const cs = getComputedStyle(document.documentElement);
    const out: Record<string, string> = {};
    for (const t of ALL_TOKENS) {
      out[t] = cs.getPropertyValue(`--color-${t}`).trim();
    }
    setHex(out);
  }, []);

  useEffect(() => {
    read();
    if (theme !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    mq.addEventListener("change", read);
    return () => mq.removeEventListener("change", read);
  }, [theme, read]);

  return hex;
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="mt-12">
      <h2 className="mb-4 font-display text-2xl leading-none">{title}</h2>
      {children}
    </section>
  );
}

function Swatches({
  label,
  tokens,
  hex,
}: {
  label: string;
  tokens: ReadonlyArray<readonly [string, string]>;
  hex: Record<string, string>;
}) {
  return (
    <>
      <h3 className="mb-2 mt-5 font-label text-[10px] uppercase tracking-[0.1em] text-muted first:mt-0">
        {label}
      </h3>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-5">
        {tokens.map(([token, use]) => (
          <div key={token} className="border border-rule bg-surface">
            <span
              className="block h-10"
              style={{ background: `var(--color-${token})` }}
            />
            <div className="px-2 py-2">
              <div className="font-label text-[10px]">{token}</div>
              <div className="mt-1 font-mono text-[10px] uppercase text-muted">
                {hex[token] || " "}
              </div>
              <div className="mt-1 text-xs leading-tight text-muted">{use}</div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

function Demo({
  label,
  note,
  children,
}: {
  label: string;
  note?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="bg-surface-2 p-4">
      <div className="mb-3 font-label text-[10px] uppercase tracking-[0.1em] text-muted">
        {label}
      </div>
      {children}
      {note && <p className="mt-3 text-xs text-muted">{note}</p>}
    </div>
  );
}

export default function StyleguidePage() {
  const [theme, setTheme] = useTheme();
  const hex = useTokenHex(theme);
  const [name, setName] = useState("Chayut A.");

  const nameError = validateDisplayName(name);

  return (
    <>
      <Header user={{ displayName: "Chayut A." }} />

      <main className="mx-auto w-full max-w-4xl px-4 pb-16 pt-8">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="font-display text-4xl leading-none">Styleguide</h1>
            <p className="mt-2 text-sm text-muted">
              Live components. Rules in <code>DESIGN_SYSTEM.md</code>.
            </p>
          </div>

          <div>
            <div className="mb-2 font-label text-[10px] uppercase tracking-[0.1em] text-muted">
              Theme
            </div>
            <div className="flex gap-2">
              {THEMES.map((t) => (
                <PixelButton
                  key={t}
                  variant={theme === t ? "primary" : "ghost"}
                  aria-pressed={theme === t}
                  onClick={() => setTheme(t)}
                >
                  {t}
                </PixelButton>
              ))}
            </div>
          </div>
        </div>

        <Section title="Colour">
          <Swatches label="Assigned" tokens={ASSIGNED} hex={hex} />
          <Swatches label="Available" tokens={AVAILABLE} hex={hex} />
          <Swatches label="Chrome, theme-aware" tokens={CHROME} hex={hex} />
          <p className="mt-3 text-xs text-muted">
            Only the chrome row changes with the theme. Assigned and available
            hold in both.
          </p>
        </Section>

        <Section title="Type">
          <div className="divide-y divide-rule border-y border-rule">
            {TYPE_ROLES.map(([role, face, cls, size]) => (
              <div
                key={face}
                className="grid gap-1 py-3 sm:grid-cols-[150px_1fr] sm:items-center sm:gap-4"
              >
                <div className="text-xs text-muted">
                  <span className="block font-label text-[10px] uppercase tracking-[0.1em] text-fg">
                    {role}
                  </span>
                  {face}
                </div>
                <div className={`${cls} ${size} leading-tight`}>
                  Fisher Timer 18h 40m
                </div>
              </div>
            ))}
          </div>
        </Section>

        <Section title="Components">
          <div className="grid gap-2 sm:grid-cols-2">
            <Demo label="Button" note="Press: cap drops, thickness gone.">
              <div className="flex flex-col items-start gap-2">
                <PixelButton>Continue</PixelButton>
                <PixelButton variant="ghost">Cancel</PixelButton>
                <PixelButton disabled>Saving…</PixelButton>
              </div>
            </Demo>

            <Demo label="Input" note="Clear the field to see the error state.">
              <PixelInput
                label="Display name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                error={nameError}
                hint="Shown in study rooms."
                maxLength={40}
              />
            </Demo>

            <Demo label="Alert, badge">
              <div className="flex flex-col items-start gap-2">
                <PixelAlert>Sign-in was cancelled.</PixelAlert>
                <PixelAlert tone="warn">
                  Couldn&rsquo;t save. Try again.
                </PixelAlert>
                <PixelBadge>BANNED</PixelBadge>
              </div>
            </Demo>

            <Demo label="Panel, tile">
              <PixelPanel>
                <div className="font-display text-lg leading-none">Panel</div>
              </PixelPanel>
              <div className="mt-2 grid grid-cols-3 gap-2">
                <StatTile value="24" caption="Sessions" />
                <StatTile value="18h" caption="Focus" />
                <StatTile value="132" caption="Fish" />
              </div>
            </Demo>
          </div>
        </Section>

        <Section title="Scene">
          <ParallaxScene layers={LAKE_LAYERS} className="h-56" />
          <p className="mt-3 text-xs text-muted">
            Nine tiling layers, one keyframe. Speed is set per layer via
            <code>animation-duration</code>; farther layers move slower.
          </p>
        </Section>

        <Section title="Grid">
          <div className="flex flex-wrap items-end gap-4 bg-surface-2 p-4">
            {[2, 3, 5].map((px) => (
              <div
                key={px}
                className="flex flex-col gap-2"
                style={{ "--px": `${px}px` } as React.CSSProperties}
              >
                <span className="font-label text-[10px] uppercase tracking-[0.1em] text-muted">
                  --px {px}
                </span>
                <PixelButton>Continue</PixelButton>
              </div>
            ))}
          </div>
          <p className="mt-3 text-xs text-muted">
            Same component. Borders, bevels, padding and focus all derive from{" "}
            <code>--px</code>.
          </p>
        </Section>
      </main>
    </>
  );
}
