"use client";

import { useEffect, useRef } from "react";
import { Game, AUTO, Scale } from "phaser";
import { MainScene } from "@/game-engine/scenes/MainScene";
import { useAuthStore } from "@/store/authStore";

export default function PhaserGame() {
  const gameRef = useRef<Game | null>(null);
  const parentEl = useRef<HTMLDivElement>(null);
  const token = useAuthStore((state) => state.token);

  useEffect(() => {
    if (!parentEl.current || !token) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: AUTO,
      parent: parentEl.current,
      width: window.innerWidth,
      height: window.innerHeight,
      backgroundColor: "#000000",
      pixelArt: true, // Crucial for retro look
      physics: {
        default: "arcade",
        arcade: { debug: true },
      },
      scene: [MainScene],
      scale: {
        mode: Scale.RESIZE, // Auto-resize with window
        autoCenter: Scale.CENTER_BOTH,
      },
    };

    gameRef.current = new Game(config);

    return () => {
      gameRef.current?.destroy(true);
      gameRef.current = null;
    };
  }, [token]);

  return <div ref={parentEl} className="w-full h-full" />;
}
