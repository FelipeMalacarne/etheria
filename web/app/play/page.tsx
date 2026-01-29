"use client";

import dynamic from "next/dynamic";
import { useGameStore } from "@/store/gameStore";
import { Card } from "@/components/ui/8bit/card";
import PlayerProfileCard from "@/components/ui/8bit/blocks/player-profile-card";
import LoadingScreen from "@/components/ui/8bit/blocks/loading-screen";

// Lazy load Phaser (Disable SSR)
const PhaserGame = dynamic(() => import("@/components/game/PhaserGame"), {
  ssr: false,
  loading: () => (
    <LoadingScreen
      className="fixed inset-0 z-50 flex items-center justify-center bg-background"
      autoProgress
      autoProgressDuration={100}
      tips={[
        "Gathering resources...",
        "Building world...",
        "Spawning creatures...",
      ]}
    />
  ),
});

export default function PlayPage() {
  const { hp, maxHp } = useGameStore();

  return (
    <div className="relative w-screen h-screen bg-black overflow-hidden">
      {/* LAYER 1: The Game Engine */}
      <div className="absolute inset-0 z-0">
        <PhaserGame />
      </div>

      {/* LAYER 2: Floating HUD (Pointer Events Trick) */}
      <div className="absolute inset-0 z-10 pointer-events-none p-6 flex flex-col justify-between">
        <PlayerProfileCard
          className="pointer-events-auto w-100 opacity-90"
          playerName="Cobalt"
          avatarSrc="/avatars/orcdev.jpeg"
          avatarFallback="C"
          level={25}
          stats={{
            health: { current: hp, max: maxHp },
            mana: { current: 320, max: 400 },
            experience: { current: 7500, max: 10000 },
          }}
          playerClass="Web Dev Warrior"
          customStats={[
            {
              label: "Stamina",
              value: 72,
              max: 100,
              color: "bg-green-500",
            },
          ]}
        />

        {/* Bottom Left: Chat */}
        <Card className="h-32 opacity-90 w-5xl pointer-events-auto">
          <p className="text-yellow-400">[System]: Welcome to RuneWeb.</p>
          <p className="text-white">Player1: selling lobbies 150gp</p>
        </Card>
      </div>
    </div>
  );
}
