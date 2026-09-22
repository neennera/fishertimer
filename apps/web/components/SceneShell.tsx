import { ReactNode } from "react";
import { Header } from "./Header";
import { ParallaxLayer, ParallaxScene } from "./ui/ParallaxScene";

export interface SceneShellProps {
  layers: ParallaxLayer[];
  children: ReactNode;
}

/**
 * The full-bleed scene background + header + centred content column shared
 * by every auth screen (/signin, /welcome, ...). Pull a new screen's panel
 * into this instead of re-laying it out — the background is a page-level
 * concern, not something each screen should re-solve.
 */
export function SceneShell({ layers, children }: SceneShellProps) {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Fixed to the viewport by its own CSS — not sized by this layout. */}
      <ParallaxScene layers={layers} />

      <Header />

      <div className="flex flex-1 items-center justify-center px-4 py-12">
        {children}
      </div>
    </div>
  );
}
