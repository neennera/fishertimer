import type { ParallaxLayer } from '../../components/ui/ParallaxScene';
import { PARALLAX_LAKE_GROUND_LAYERS } from './parallax-lake';
import { SKY_LAYERS } from './sky';

// The full /signin background: the sky/cloud "cover" layers behind the
// parallax-lake "tile" ground layers. Shared with the styleguide's "Scene"
// demo so the two never drift apart.
export const SIGNIN_SCENE_LAYERS: ParallaxLayer[] = [...SKY_LAYERS, ...PARALLAX_LAKE_GROUND_LAYERS];
