import { PixelBadge } from '../../../components/ui/PixelBadge';

export function StatusBadge({ active }: { active: boolean }) {
  return (
    <PixelBadge className={active ? 'pixel-badge--ok' : 'pixel-badge--muted'}>
      {active ? 'ACTIVE' : 'ENDED'}
    </PixelBadge>
  );
}
