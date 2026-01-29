import { Scene } from "phaser";
import { NetworkClient } from "@/game-engine/network/client";
import { PlayerState, Welcome } from "@/game-engine/network/packets";
import { gameStoreApi } from "@/store/gameStore";

export class MainScene extends Scene {
  private network: NetworkClient | null = null;
  private playerSprites = new Map<string, Phaser.GameObjects.Rectangle>();
  private localPlayerId: string | null = null;
  private isFollowing = false;
  private playerSize = 18;
  private isShuttingDown = false;

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
    this.playerSize = tileSize * 0.6;

    this.setupNetwork();

    this.input.on("pointerdown", (pointer: Phaser.Input.Pointer) => {
      const worldPoint = this.cameras.main.getWorldPoint(pointer.x, pointer.y);
      this.network?.sendMoveIntent(
        Math.round(worldPoint.x),
        Math.round(worldPoint.y),
      );
    });

    this.events.once("shutdown", () => {
      this.isShuttingDown = true;
      this.network?.disconnect();
      this.playerSprites.clear();
    });
  }

  private setupNetwork() {
    this.network = new NetworkClient({
      onConnectionChange: (connected) => {
        if (this.isShuttingDown) {
          return;
        }
        gameStoreApi.getState().setConnected(connected);
        if (!connected) {
          gameStoreApi.getState().setPlayerId(null);
          this.localPlayerId = null;
          this.isFollowing = false;
        }
      },
      onWelcome: (welcome) => {
        this.handleWelcome(welcome);
      },
      onStateUpdate: (state) => {
        this.syncPlayers(state.players);
      },
    });

    this.network.connect(this.getWebSocketUrl());
  }

  private handleWelcome(welcome: Welcome) {
    if (this.isShuttingDown || !this.sys.isActive()) {
      return;
    }

    this.localPlayerId = welcome.id;
    gameStoreApi.getState().setPlayerId(welcome.id);
    const sprite = this.playerSprites.get(welcome.id);
    if (sprite) {
      sprite.setFillStyle(0xff6b6b);
    }

    this.followPlayerIfReady();
  }

  private syncPlayers(players: PlayerState[]) {
    if (this.isShuttingDown || !this.sys.isActive()) {
      return;
    }

    const seen = new Set(players.map((player) => player.id));

    for (const [id, sprite] of this.playerSprites) {
      if (!seen.has(id)) {
        sprite.destroy();
        this.playerSprites.delete(id);
        if (id === this.localPlayerId) {
          this.isFollowing = false;
        }
      }
    }

    for (const player of players) {
      const isLocal = player.id === this.localPlayerId;
      let sprite = this.playerSprites.get(player.id);

      if (!sprite) {
        sprite = this.add.rectangle(
          player.x,
          player.y,
          this.playerSize,
          this.playerSize,
          isLocal ? 0xff6b6b : 0x4d96ff,
        );
        this.playerSprites.set(player.id, sprite);
      } else {
        sprite.setPosition(player.x, player.y);
        if (isLocal) {
          sprite.setFillStyle(0xff6b6b);
        }
      }
    }

    this.followPlayerIfReady();
  }

  private followPlayerIfReady() {
    if (this.isFollowing || this.playerSprites.size === 0) {
      return;
    }

    const targetId = this.localPlayerId ?? this.playerSprites.keys().next().value;
    if (!targetId) {
      return;
    }

    const sprite = this.playerSprites.get(targetId);
    if (!sprite) {
      return;
    }

    this.cameras.main.startFollow(sprite);
    this.isFollowing = true;
  }

  private getWebSocketUrl() {
    const envUrl = process.env.NEXT_PUBLIC_WS_URL;
    if (envUrl) {
      return envUrl;
    }

    const protocol = window.location.protocol === "https:" ? "wss" : "ws";
    return `${protocol}://${window.location.hostname}:8080/ws`;
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
