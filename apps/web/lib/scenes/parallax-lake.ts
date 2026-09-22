import type { ParallaxLayer } from '../../components/ui/ParallaxScene';

// Farthest to nearest, slowest to fastest. Pixel-art sprites (384x216) — see
// pixel.css's .pixel-parallax__layer--tile for why these tile at whole
// multiples of --px instead of scaling like the sky/cloud layers do. Sky and
// clouds for this scene come from lib/scenes/sky.ts instead of this set's
// own sky.png/clouds.png, which are unused now.
export const PARALLAX_LAKE_GROUND_LAYERS: ParallaxLayer[] = [
  { src: '/sprites/scene/parallax-lake/mountains.png', fit: 'tile', speed: 0.2 },
  { src: '/sprites/scene/parallax-lake/forest-far.png', fit: 'tile', speed: 0.35 },
  { src: '/sprites/scene/parallax-lake/forest-mid.png', fit: 'tile', speed: 0.5 },
  { src: '/sprites/scene/parallax-lake/forest-near.png', fit: 'tile', speed: 0.7 },
  { src: '/sprites/scene/parallax-lake/valley-fill.png', fit: 'tile', speed: 0.85 },
  { src: '/sprites/scene/parallax-lake/foreground.png', fit: 'tile', speed: 1 },
  { src: '/sprites/scene/parallax-lake/water.png', fit: 'tile', speed: 1.2 },
];
