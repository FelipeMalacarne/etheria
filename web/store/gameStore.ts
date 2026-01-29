import { create } from "zustand";

interface GameState {
  isConnected: boolean;
  playerId: string | null;
  hp: number;
  maxHp: number;
  inventoryOpen: boolean;

  // Actions
  setConnected: (status: boolean) => void;
  setPlayerId: (id: string | null) => void;
  updateHp: (current: number, max: number) => void;
  toggleInventory: () => void;
}

export const useGameStore = create<GameState>((set) => ({
  isConnected: false,
  playerId: null,
  hp: 10,
  maxHp: 10,
  inventoryOpen: false,

  setConnected: (status) => set({ isConnected: status }),
  setPlayerId: (id) => set({ playerId: id }),
  updateHp: (current, max) => set({ hp: current, maxHp: max }),
  toggleInventory: () =>
    set((state) => ({ inventoryOpen: !state.inventoryOpen })),
}));

// EXPORT API FOR PHASER
// We export this so non-React files can read/write to the store
export const gameStoreApi = useGameStore;
