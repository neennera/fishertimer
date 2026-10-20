import type { ReactNode } from 'react';
import { RequireAdmin } from '../../components/RequireAdmin';
import { SessionHeader } from '../../components/SessionHeader';
import { ParallaxScene } from '../../components/ui/ParallaxScene';
import { ADMIN_SCENE_LAYERS } from '../../lib/scenes/admin-scene';

// Shared by every admin page (like app/rooms/layout.tsx), so the scene and
// header persist while moving between them; only the panels swap.
export default function AdminLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={ADMIN_SCENE_LAYERS} />
      <SessionHeader />
      <RequireAdmin>{children}</RequireAdmin>
    </div>
  );
}
