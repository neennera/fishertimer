import type { ReactNode } from "react";
import { SessionHeader } from "../../components/SessionHeader";
import { ParallaxScene } from "../../components/ui/ParallaxScene";
import { SIGNIN_SCENE_LAYERS } from "../../lib/scenes/signin-scene";

// Shared by the room list and every room, so the lake behind them and the
// header persist while moving between the two; only the panels swap.
export default function RoomsLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <ParallaxScene layers={SIGNIN_SCENE_LAYERS} />
      <SessionHeader />
      {children}
    </div>
  );
}
