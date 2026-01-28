Here is your complete **Frontend Setup Documentation**. This guide focuses on getting the **Next.js + Phaser + Retro UI** architecture running immediately.

Save this as `FRONTEND_SETUP.md` in your project root.

---

# ⚔️ RuneWeb: Frontend Documentation

## 1. Project Initialization

We will use **Next.js 14+ (App Router)** with **TypeScript** and **Tailwind CSS**.

```bash
# 1. Create the Next.js project
npx create-next-app@latest runeweb-client
# Select the following options:
# - TypeScript: Yes
# - ESLint: Yes
# - Tailwind CSS: Yes
# - `src/` directory: Yes
# - App Router: Yes
# - Import alias: @/*

# 2. Enter directory
cd runeweb-client

# 3. Install Game & State Dependencies
npm install phaser zustand

# 4. Install Utility Class Merger (Optional but recommended for UI components)
npm install clsx tailwind-merge

```

---

## 2. Directory Structure

We strictly separate the **React UI** from the **Phaser Game Engine**.

```text
src/
├── app/
│   ├── layout.tsx         # Global font setup (Pixel Font)
│   └── play/
│       └── page.tsx       # The Game Entry Point (Canvas + HUD)
├── components/
│   ├── game/
│   │   └── PhaserGame.tsx # The Wrapper (React <-> Phaser)
│   └── ui/
│       ├── RetroCard.tsx  # 8-bit styled container
│       └── GameHUD.tsx    # The floating interface
├── game-engine/           # PURE PHASER LOGIC (No React here)
│   ├── game.ts            # Game Config
│   └── scenes/
│       ├── BootScene.ts   # Asset Loading
│       └── MainScene.ts   # Gameplay Logic
└── store/
    └── gameStore.ts       # The Bridge (Zustand)

```

---

## 3. The "Bridge" (State Management)

This file allows Phaser to update React (e.g., HP changes) without re-rendering the whole game.

**File:** `src/store/gameStore.ts`

```typescript
import { create } from "zustand";

interface GameState {
  isConnected: boolean;
  hp: number;
  maxHp: number;
  inventoryOpen: boolean;

  // Actions
  setConnected: (status: boolean) => void;
  updateHp: (current: number, max: number) => void;
  toggleInventory: () => void;
}

export const useGameStore = create<GameState>((set) => ({
  isConnected: false,
  hp: 10,
  maxHp: 10,
  inventoryOpen: false,

  setConnected: (status) => set({ isConnected: status }),
  updateHp: (current, max) => set({ hp: current, maxHp: max }),
  toggleInventory: () =>
    set((state) => ({ inventoryOpen: !state.inventoryOpen })),
}));

// EXPORT API FOR PHASER
// We export this so non-React files can read/write to the store
export const gameStoreApi = useGameStore;
```

---

## 4. The Retro UI Components

We create a reusable "Card" component that applies the 8-bit style using pure CSS Borders (so it's lightweight).

**File:** `src/components/ui/RetroCard.tsx`

```tsx
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

interface RetroCardProps {
  children: React.ReactNode;
  className?: string;
}

export function RetroCard({ children, className }: RetroCardProps) {
  return (
    <div
      className={twMerge(
        "bg-[#5d5d5d] p-4 font-mono text-white relative",
        className,
      )}
      style={{
        // The Classic "3D Bevel" Effect
        borderTop: "4px solid #a0a0a0",
        borderLeft: "4px solid #a0a0a0",
        borderRight: "4px solid #2b2b2b",
        borderBottom: "4px solid #2b2b2b",
        boxShadow: "2px 2px 0px 0px #000000",
      }}
    >
      {children}
    </div>
  );
}
```

---

## 5. The Phaser Setup

### A. The Main Scene

**File:** `src/game-engine/scenes/MainScene.ts`

```typescript
import { Scene } from "phaser";
import { gameStoreApi } from "@/store/gameStore";

export class MainScene extends Scene {
  constructor() {
    super("MainScene");
  }

  create() {
    // 1. Create a simple player (Red Square)
    const player = this.add.rectangle(400, 300, 32, 32, 0xff0000);
    this.physics.add.existing(player);

    // 2. Simple interaction example
    this.input.on("pointerdown", () => {
      // Simulate taking damage when clicking
      const currentHp = gameStoreApi.getState().hp;
      const newHp = Math.max(0, currentHp - 1);

      // Update React Store
      gameStoreApi.getState().updateHp(newHp, 10);
    });

    // 3. Add some text
    this.add.text(10, 10, "Click to take dmg", {
      fontSize: "16px",
      color: "#ffffff",
    });
  }
}
```

### B. The React Wrapper

**File:** `src/components/game/PhaserGame.tsx`

```tsx
"use client";

import { useEffect, useRef } from "react";
import { Game, AUTO, Scale } from "phaser";
import { MainScene } from "@/game-engine/scenes/MainScene";

export default function PhaserGame() {
  const gameRef = useRef<Game | null>(null);
  const parentEl = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!parentEl.current) return;

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
    };
  }, []);

  return <div ref={parentEl} className="w-full h-full" />;
}
```

---

## 6. The Page Layout (Putting it Together)

**File:** `src/app/play/page.tsx`

```tsx
"use client";

import dynamic from "next/dynamic";
import { useGameStore } from "@/store/gameStore";
import { RetroCard } from "@/components/ui/RetroCard";

// Lazy load Phaser (Disable SSR)
const PhaserGame = dynamic(() => import("@/components/game/PhaserGame"), {
  ssr: false,
  loading: () => (
    <div className="bg-black h-screen text-white p-10">Loading Gielinor...</div>
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
        {/* Top Left: HP Bar */}
        <div className="pointer-events-auto w-64">
          <RetroCard>
            <div className="flex justify-between mb-2">
              <span>Hitpoints</span>
              <span>
                {hp} / {maxHp}
              </span>
            </div>
            {/* Health Bar Container */}
            <div className="w-full h-4 bg-red-900 border border-black">
              <div
                className="h-full bg-green-500 transition-all duration-300"
                style={{ width: `${(hp / maxHp) * 100}%` }}
              />
            </div>
          </RetroCard>
        </div>

        {/* Bottom Left: Chat */}
        <div className="pointer-events-auto w-96">
          <RetroCard className="h-32 opacity-90">
            <p className="text-yellow-400">[System]: Welcome to RuneWeb.</p>
            <p className="text-white">Player1: selling lobbies 150gp</p>
          </RetroCard>
        </div>
      </div>
    </div>
  );
}
```

---

## 7. Adding the Font

To make it look truly "Retro", add a pixel font.

1. Download a font like **"Press Start 2P"** or **"VT323"** (Google Fonts).
2. Add it to `src/app/layout.tsx`:

```tsx
import { VT323 } from "next/font/google";
import "./globals.css";

const pixelFont = VT323({
  weight: "400",
  subsets: ["latin"],
  variable: "--font-pixel",
});

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={`${pixelFont.variable} font-pixel antialiased`}>
        {children}
      </body>
    </html>
  );
}
```

3. Update `tailwind.config.ts`:

```ts
theme: {
  extend: {
    fontFamily: {
      pixel: ['var(--font-pixel)', 'monospace'],
    },
  },
},

```

## Next Steps

Run `npm run dev` and navigate to `http://localhost:3000/play`.
You should see:

1. A black screen (Phaser Canvas).
2. A Red Square (Player).
3. A floating "Retro Style" HP bar.
4. Clicking the screen lowers the HP bar (demonstrating the Bridge is working).
