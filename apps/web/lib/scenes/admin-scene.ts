import type { ParallaxLayer } from '../../components/ui/ParallaxScene';

// The /admin background: an alpine snow-mountain scene, deliberately unlike
// /signin's lake. Back to front, all "cover" layers from the skies set — the
// mountains sit still, the clouds drift (a leg each way, see
// .pixel-parallax__layer--drift).
export const ADMIN_SCENE_LAYERS: ParallaxLayer[] = [
  { src: '/sprites/scene/skies/Sky_sky.png', fit: 'cover' },
  { src: '/sprites/scene/skies/sky_clouds.png', fit: 'cover', driftSeconds: 120 },
  { src: '/sprites/scene/skies/Sky_back_mountain.png', fit: 'cover' },
  { src: '/sprites/scene/skies/sky_front_mountain.png', fit: 'cover' },
  { src: '/sprites/scene/skies/Sky_front_cloud.png', fit: 'cover', driftSeconds: 70 },
];
