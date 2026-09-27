// Art per catalog item id; schema.dbml's reward_items has no image column.
// Sprites: 16x16, facing right, or a 48x16 strip of 3 swim frames.

export type FishPlaceholderColor = 'rust' | 'sky-fill' | 'amber' | 'sage' | 'cream';

export interface FishSprite {
  /** Unset = placeholder shape. */
  src?: string;
  frames?: 1 | 3;
  placeholder: FishPlaceholderColor;
}

const FISH_SPRITES: Record<string, FishSprite> = {
  'fish-clownfish': { src: '/sprites/fish/Clownfish.png', placeholder: 'rust' },
  // A blue tang is a surgeonfish; the sprite pack names it by family.
  'fish-blue-tang': { src: '/sprites/fish/Surgeonfish.png', placeholder: 'sky-fill' },
  'fish-goldfish': { src: '/sprites/fish/Goldfish.png', placeholder: 'amber' },
  'fish-pufferfish': { src: '/sprites/fish/Pufferfish.png', placeholder: 'amber' },
  'fish-angelfish': { src: '/sprites/fish/Angelfish.png', placeholder: 'cream' },
  'fish-anchovy': { src: '/sprites/fish/Anchovy.png', placeholder: 'sky-fill' },
  'fish-bass': { src: '/sprites/fish/Bass.png', placeholder: 'sage' },
  'fish-catfish': { src: '/sprites/fish/Catfish.png', placeholder: 'sage' },
  'fish-rainbow-trout': { src: '/sprites/fish/Rainbow Trout.png', placeholder: 'sage' },
};

const UNKNOWN_FISH: FishSprite = { placeholder: 'sage' };

export function fishSprite(itemId: string): FishSprite {
  return FISH_SPRITES[itemId] ?? UNKNOWN_FISH;
}
