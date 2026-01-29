import { Scene } from "phaser";
import { gameStoreApi } from "@/store/gameStore";

export class MainScene extends Scene {
  constructor() {
    super("MainScene");
  }

  create() {
    // 1. Create a simple player (Red Square)
    const player = this.add.rectangle(800, 800, 32, 32, 0xff0000);
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
    // this.add.text(10, 10, "Click to take dmg", {
    //   fontSize: "16px",
    //   color: "#ffffff",
    // });
  }
}
