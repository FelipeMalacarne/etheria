import { Scene } from "phaser";
import { gameStoreApi } from "@/store/gameStore";

export class MainScene extends Scene {
  constructor() {
    super("MainScene");
  }

  create() {
    const tileSize = 32;
    const mapWidth = 50;
    const mapHeight = 50;
    const tilesKey = "basic-tiles";

    this.ensureTilesetTexture(tilesKey, tileSize);

    const map = this.make.tilemap({
      data: this.buildMapData(mapWidth, mapHeight),
      tileWidth: tileSize,
      tileHeight: tileSize,
    });

    const tileset = map.addTilesetImage(
      tilesKey,
      tilesKey,
      tileSize,
      tileSize,
      0,
      0,
    );
    map.createLayer(0, tileset, 0, 0);

    this.cameras.main.setBounds(0, 0, map.widthInPixels, map.heightInPixels);

    const player = this.add.rectangle(
      map.widthInPixels / 2,
      map.heightInPixels / 2,
      tileSize * 0.6,
      tileSize * 0.6,
      0xff0000,
    );
    this.physics.add.existing(player);
    this.cameras.main.startFollow(player);

    // 2. Simple interaction example
    this.input.on("pointerdown", () => {
      // Simulate taking damage when clicking
      const currentHp = gameStoreApi.getState().hp;
      const newHp = Math.max(0, currentHp - 1);

      // Update React Store
      gameStoreApi.getState().updateHp(newHp, 10);
    });
  }

  private ensureTilesetTexture(key: string, tileSize: number) {
    if (this.textures.exists(key)) return;

    const texture = this.textures.createCanvas(key, tileSize * 3, tileSize);
    const ctx = texture.getContext();

    ctx.fillStyle = "#2f6f3e";
    ctx.fillRect(0, 0, tileSize, tileSize);

    ctx.fillStyle = "#8b5a2b";
    ctx.fillRect(tileSize, 0, tileSize, tileSize);

    ctx.fillStyle = "#2b5a8b";
    ctx.fillRect(tileSize * 2, 0, tileSize, tileSize);

    texture.refresh();
  }

  private buildMapData(width: number, height: number) {
    const data: number[][] = [];

    for (let y = 0; y < height; y += 1) {
      const row: number[] = [];
      for (let x = 0; x < width; x += 1) {
        const isBorder =
          x === 0 || y === 0 || x === width - 1 || y === height - 1;
        let tileIndex = 0;

        if (isBorder) {
          tileIndex = 2;
        } else if ((x + y) % 7 === 0) {
          tileIndex = 1;
        }

        row.push(tileIndex);
      }
      data.push(row);
    }

    return data;
  }
}
