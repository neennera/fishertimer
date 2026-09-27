import type { ParallaxLayer } from '../../components/ui/ParallaxScene';

// Back to front. Smooth painterly/gradient art — see pixel.css's
// .pixel-parallax__layer--cover for why these use "cover", not tiling. The
// sky itself stays static; the clouds drift slowly (a 90s leg each way, 3
// minutes round trip) — see .pixel-parallax__layer--drift.
export const SKY_LAYERS: ParallaxLayer[] = [
  { src: '/sprites/scene/skies/Sky_sky.png', fit: 'cover' },
  { src: '/sprites/scene/skies/sky_clouds.png', fit: 'cover', driftSeconds: 90 },
];
