"use client";

import dynamic from "next/dynamic";
import { useGameStore } from "@/store/gameStore";
import { Card } from "@/components/ui/8bit/card";
import PlayerProfileCard from "@/components/ui/8bit/blocks/player-profile-card";
import LoadingScreen from "@/components/ui/8bit/blocks/loading-screen";
import FriendList from "@/components/ui/8bit/blocks/friend-list";
import { usePlayers } from "@/hooks/use-players";

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
  const { hp, maxHp, isConnected, playerId } = useGameStore();
  const { players } = usePlayers();
  const friendPlayers = players.map((player) => ({
    id: player.id,
    name: player.id === playerId ? `You (${player.id})` : `Player ${player.id}`,
    status: player.id === playerId ? "ingame" : "online",
    avatarFallback: player.id.charAt(0).toUpperCase(),
    activity: `(${Math.round(player.x)}, ${Math.round(player.y)})`,
  }));

  return (
    <div className="relative w-screen h-screen bg-black overflow-hidden">
      {/* LAYER 1: The Game Engine */}
      <div className="absolute inset-0 z-0">
        <PhaserGame />
      </div>

      {/* LAYER 2: Floating HUD (Pointer Events Trick) */}
      <div className="absolute inset-0 z-10 pointer-events-none p-6 flex flex-col justify-between">
        <div className="flex flex-col gap-3 pointer-events-auto w-100 opacity-90">
          <PlayerProfileCard
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
          <Card className="px-4 py-2 text-xs">
            <div className="flex items-center justify-between gap-4">
              <span className="uppercase tracking-[0.2em] text-muted-foreground">
                Network
              </span>
              <span className={isConnected ? "text-green-400" : "text-red-400"}>
                {isConnected ? "Connected" : "Disconnected"}
              </span>
            </div>
            <div className="mt-1 text-muted-foreground">
              Player ID: {playerId ?? "—"}
            </div>
          </Card>
        </div>

        {/* Bottom Left: Chat */}
        <Card className="h-32 opacity-90 w-5xl pointer-events-auto">
          <p className="text-yellow-400">[System]: Welcome to RuneWeb.</p>
          <p className="text-white">Player1: selling lobbies 150gp</p>
        </Card>
      </div>

      <div className="absolute right-6 top-6 z-10 pointer-events-auto hidden lg:block w-72">
        <FriendList players={friendPlayers} />
      </div>
    </div>
  );
}
