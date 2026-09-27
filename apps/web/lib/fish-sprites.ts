// Sprite for a caught fish: by asset_url file name, then by name, else the
// placeholder. Sprites are 16x16 facing right, or a 48x16 strip of 3 frames.

export type FishPlaceholderColor = 'rust' | 'sky-fill' | 'amber' | 'sage' | 'cream';

export interface FishSprite {
  /** Unset = placeholder shape. */
  src?: string;
  frames?: 1 | 3;
  placeholder: FishPlaceholderColor;
}

const KNOWN: { file: string; names: string[]; placeholder: FishPlaceholderColor }[] = [
  { file: 'Clownfish.png', names: ['clownfish'], placeholder: 'rust' },
  // A blue tang is a surgeonfish; the sprite pack names it by family.
  { file: 'Surgeonfish.png', names: ['surgeonfish', 'blue tang'], placeholder: 'sky-fill' },
  { file: 'Goldfish.png', names: ['goldfish'], placeholder: 'amber' },
  { file: 'Pufferfish.png', names: ['pufferfish'], placeholder: 'amber' },
  { file: 'Angelfish.png', names: ['angelfish'], placeholder: 'cream' },
  { file: 'Anchovy.png', names: ['anchovy'], placeholder: 'sky-fill' },
  { file: 'Bass.png', names: ['bass'], placeholder: 'sage' },
  { file: 'Catfish.png', names: ['catfish'], placeholder: 'sage' },
  { file: 'Rainbow Trout.png', names: ['rainbow trout'], placeholder: 'sage' },
];

const PLACEHOLDER_COLORS: FishPlaceholderColor[] = ['rust', 'sky-fill', 'amber', 'sage', 'cream'];

function fileName(url: string): string {
  const last = url.split(/[?#]/)[0]!.split('/').pop() ?? '';
  try {
    return decodeURIComponent(last).toLowerCase();
  } catch {
    return last.toLowerCase();
  }
}

export function fishSprite({ name, asset_url }: { name: string; asset_url: string }): FishSprite {
  const file = fileName(asset_url);
  const key = name.trim().toLowerCase();
  const known =
    (file && KNOWN.find((k) => k.file.toLowerCase() === file)) || KNOWN.find((k) => k.names.includes(key));
  if (known) {
    return { src: `/sprites/fish/${known.file}`, placeholder: known.placeholder };
  }
  // Colour from the name, so it's stable.
  const hash = [...key].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7);
  return { placeholder: PLACEHOLDER_COLORS[hash % PLACEHOLDER_COLORS.length]! };
}
