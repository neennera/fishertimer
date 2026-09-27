// Fish tank background, back to front. `front` layers draw over all but the
// nearest fish.

export interface TankLayer {
  src: string;
  /** Native width; all layers are 192 tall. */
  width: number;
  /** Art pixels to shift the layer left. */
  offset?: number;
  front?: boolean;
}

export const FISHTANK_LAYERS: TankLayer[] = [
  { src: '/sprites/scene/parallax-fishtank/far.png', width: 256 },
  { src: '/sprites/scene/parallax-fishtank/sand.png', width: 256 },
  // Offset puts the tall rock at the left edge, leaving open water.
  {
    src: '/sprites/scene/parallax-fishtank/foregound-merged.png',
    width: 512,
    offset: 60,
    front: true,
  },
];
