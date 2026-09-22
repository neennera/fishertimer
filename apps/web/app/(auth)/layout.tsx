import type { ReactNode } from "react";
import { SceneShell } from "../../components/SceneShell";
import { SIGNIN_SCENE_LAYERS } from "../../lib/scenes/signin-scene";

// Shared by every screen in the sign-in -> first-time-setup flow. A route
// group (the URLs stay /signin and /welcome) so this layout persists across
// navigation between them instead of each page re-rendering its own
// SceneShell — the background and header don't remount, only the panel
// (children) swaps.
export default function AuthLayout({ children }: { children: ReactNode }) {
  return <SceneShell layers={SIGNIN_SCENE_LAYERS}>{children}</SceneShell>;
}
