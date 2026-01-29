import { Scene } from "phaser";
import { NetworkClient } from "@/game-engine/network/client";
import {
  PlayerState,
  POSITION_SCALE,
  Welcome,
} from "@/game-engine/network/packets";
import { gameStoreApi } from "@/store/gameStore";

export class MainScene extends Scene {
  private network: NetworkClient | null = null;
  private playerSprites = new Map<string, Phaser.GameObjects.Rectangle>();
  private playerTargets = new Map<string, { x: number; y: number }>();
  private localPlayerId: string | null = null;
  private isFollowing = false;
  private playerSize = 18;
  private isShuttingDown = false;
  private tileSize = 32;
  private mapWidth = 0;
  private mapHeight = 0;
  private mapData: number[][] = [];
  private localPath: { x: number; y: number }[] = [];
  private localPathIndex = 0;
  private serverPathIndex = 0;
  private localPredicted: { x: number; y: number } | null = null;
  private localServerPos: { x: number; y: number } | null = null;
  private interpolationSpeed = 220;
  private pathTargetThreshold = 6;
  private clientMoveSpeed = 140;
  private correctionRate = 8;
  private snapDistance = 96;

  constructor() {
    super("MainScene");
  }

  create() {
    const tileSize = 32;
    const mapWidth = 50;
    const mapHeight = 50;
    const tilesKey = "basic-tiles";

    this.ensureTilesetTexture(tilesKey, tileSize);

    this.tileSize = tileSize;
    this.mapWidth = mapWidth;
    this.mapHeight = mapHeight;
    this.mapData = this.buildMapData(mapWidth, mapHeight);

    const map = this.make.tilemap({
      data: this.mapData,
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
      this.queuePathTo(worldPoint.x, worldPoint.y);
    });

    this.events.once("shutdown", () => {
      this.isShuttingDown = true;
      this.network?.disconnect();
      this.playerSprites.clear();
      this.playerTargets.clear();
      this.localPath = [];
      this.localPathIndex = 0;
      this.serverPathIndex = 0;
      this.localPredicted = null;
      this.localServerPos = null;
    });
  }

  update(_time: number, delta: number) {
    if (this.isShuttingDown) {
      return;
    }

    const deltaSeconds = delta / 1000;
    this.interpolatePlayers(deltaSeconds);
    this.updateLocalPrediction(deltaSeconds);
    this.advanceServerPath();
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
          this.localPath = [];
          this.localPathIndex = 0;
          this.serverPathIndex = 0;
          this.localPredicted = null;
          this.localServerPos = null;
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
        this.playerTargets.delete(id);
        if (id === this.localPlayerId) {
          this.isFollowing = false;
          this.localPredicted = null;
          this.localServerPos = null;
          this.localPath = [];
          this.localPathIndex = 0;
          this.serverPathIndex = 0;
        }
      }
    }

    for (const player of players) {
      const isLocal = player.id === this.localPlayerId;
      let sprite = this.playerSprites.get(player.id);
      const worldX = this.fromNetworkPosition(player.x);
      const worldY = this.fromNetworkPosition(player.y);
      this.playerTargets.set(player.id, { x: worldX, y: worldY });
      if (isLocal) {
        this.localServerPos = { x: worldX, y: worldY };
        if (!this.localPredicted) {
          this.localPredicted = { x: worldX, y: worldY };
        }
      }

      if (!sprite) {
        sprite = this.add.rectangle(
          worldX,
          worldY,
          this.playerSize,
          this.playerSize,
          isLocal ? 0xff6b6b : 0x4d96ff,
        );
        this.playerSprites.set(player.id, sprite);
        if (isLocal && this.localPredicted) {
          sprite.setPosition(this.localPredicted.x, this.localPredicted.y);
        }
      } else if (isLocal) {
        sprite.setFillStyle(0xff6b6b);
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

  private toNetworkPosition(value: number) {
    return Math.round(value * POSITION_SCALE);
  }

  private fromNetworkPosition(value: number) {
    return value / POSITION_SCALE;
  }

  private interpolatePlayers(deltaSeconds: number) {
    if (deltaSeconds <= 0) {
      return;
    }

    const step = this.interpolationSpeed * deltaSeconds;

    for (const [id, sprite] of this.playerSprites) {
      if (id === this.localPlayerId) {
        continue;
      }
      const target = this.playerTargets.get(id);
      if (!target) {
        continue;
      }

      const dx = target.x - sprite.x;
      const dy = target.y - sprite.y;
      const distance = Math.hypot(dx, dy);
      if (distance === 0) {
        continue;
      }

      if (distance <= step) {
        sprite.setPosition(target.x, target.y);
        continue;
      }

      sprite.x += (dx / distance) * step;
      sprite.y += (dy / distance) * step;
    }
  }

  private queuePathTo(worldX: number, worldY: number) {
    if (!this.network) {
      return;
    }

    const start = this.getLocalPlayerWorldPosition();
    if (!start) {
      this.network.sendMoveIntent(
        this.toNetworkPosition(worldX),
        this.toNetworkPosition(worldY),
      );
      return;
    }

    const startGrid = this.worldToGrid(start.x, start.y);
    const goalGrid = this.worldToGrid(worldX, worldY);

    const path = this.findPath(startGrid, goalGrid);
    if (path.length === 0) {
      this.network.sendMoveIntent(
        this.toNetworkPosition(worldX),
        this.toNetworkPosition(worldY),
      );
      return;
    }

    this.localPath = path
      .slice(1)
      .map((point) => this.gridToWorld(point.x, point.y));
    this.localPathIndex = 0;
    this.serverPathIndex = 0;

    this.sendNextPathTarget();
  }

  private sendNextPathTarget() {
    if (!this.network || this.serverPathIndex >= this.localPath.length) {
      return;
    }

    const next = this.localPath[this.serverPathIndex];
    this.network.sendMoveIntent(
      this.toNetworkPosition(next.x),
      this.toNetworkPosition(next.y),
    );
  }

  private advanceServerPath() {
    if (!this.network || this.localPath.length === 0) {
      return;
    }

    if (this.serverPathIndex >= this.localPath.length) {
      return;
    }

    const reference = this.localServerPos ?? this.localPredicted;
    if (!reference) {
      return;
    }

    const next = this.localPath[this.serverPathIndex];
    const distance = Phaser.Math.Distance.Between(
      reference.x,
      reference.y,
      next.x,
      next.y,
    );

    if (distance <= this.pathTargetThreshold) {
      this.serverPathIndex += 1;
      this.sendNextPathTarget();
    }
  }

  private getLocalPlayerWorldPosition() {
    if (!this.localPlayerId) {
      return null;
    }

    return (
      this.localPredicted ??
      this.localServerPos ??
      this.playerTargets.get(this.localPlayerId) ??
      this.playerSprites.get(this.localPlayerId) ??
      null
    );
  }

  private updateLocalPrediction(deltaSeconds: number) {
    if (!this.localPlayerId || deltaSeconds <= 0) {
      return;
    }

    const sprite = this.playerSprites.get(this.localPlayerId);
    if (!sprite) {
      return;
    }

    if (!this.localPredicted) {
      this.localPredicted = { x: sprite.x, y: sprite.y };
    }

    if (this.localPathIndex < this.localPath.length) {
      const target = this.localPath[this.localPathIndex];
      const dx = target.x - this.localPredicted.x;
      const dy = target.y - this.localPredicted.y;
      const distance = Math.hypot(dx, dy);
      const step = this.clientMoveSpeed * deltaSeconds;

      if (distance <= step) {
        this.localPredicted.x = target.x;
        this.localPredicted.y = target.y;
        this.localPathIndex += 1;
      } else if (distance > 0) {
        this.localPredicted.x += (dx / distance) * step;
        this.localPredicted.y += (dy / distance) * step;
      }
    }

    this.applyServerCorrection(deltaSeconds);
    sprite.setPosition(this.localPredicted.x, this.localPredicted.y);
  }

  private applyServerCorrection(deltaSeconds: number) {
    if (!this.localPredicted || !this.localServerPos) {
      return;
    }

    const dx = this.localServerPos.x - this.localPredicted.x;
    const dy = this.localServerPos.y - this.localPredicted.y;
    const distance = Math.hypot(dx, dy);

    if (distance > this.snapDistance) {
      this.localPredicted.x = this.localServerPos.x;
      this.localPredicted.y = this.localServerPos.y;
      return;
    }

    if (distance === 0) {
      return;
    }

    const correction = 1 - Math.exp(-this.correctionRate * deltaSeconds);
    this.localPredicted.x += dx * correction;
    this.localPredicted.y += dy * correction;
  }

  private worldToGrid(x: number, y: number) {
    return {
      x: Math.floor(x / this.tileSize),
      y: Math.floor(y / this.tileSize),
    };
  }

  private gridToWorld(x: number, y: number) {
    return {
      x: (x + 0.5) * this.tileSize,
      y: (y + 0.5) * this.tileSize,
    };
  }

  private isWalkable(x: number, y: number) {
    if (x < 0 || y < 0 || x >= this.mapWidth || y >= this.mapHeight) {
      return false;
    }

    return this.mapData[y][x] !== 2;
  }

  private findPath(
    start: { x: number; y: number },
    goal: { x: number; y: number },
  ) {
    if (!this.isWalkable(goal.x, goal.y)) {
      return [];
    }

    const startKey = `${start.x},${start.y}`;
    const goalKey = `${goal.x},${goal.y}`;

    const open: Array<{ x: number; y: number; f: number }> = [
      {
        x: start.x,
        y: start.y,
        f: this.heuristic(start, goal),
      },
    ];
    const openSet = new Set([startKey]);
    const cameFrom = new Map<string, string>();
    const gScore = new Map<string, number>([[startKey, 0]]);

    while (open.length > 0) {
      let currentIndex = 0;
      for (let i = 1; i < open.length; i += 1) {
        if (open[i].f < open[currentIndex].f) {
          currentIndex = i;
        }
      }

      const current = open.splice(currentIndex, 1)[0];
      const currentKey = `${current.x},${current.y}`;
      openSet.delete(currentKey);

      if (currentKey === goalKey) {
        return this.reconstructPath(cameFrom, goalKey);
      }

      const neighbors = [
        { x: current.x + 1, y: current.y },
        { x: current.x - 1, y: current.y },
        { x: current.x, y: current.y + 1 },
        { x: current.x, y: current.y - 1 },
      ];

      for (const neighbor of neighbors) {
        if (!this.isWalkable(neighbor.x, neighbor.y)) {
          continue;
        }

        const neighborKey = `${neighbor.x},${neighbor.y}`;
        const tentativeG = (gScore.get(currentKey) ?? 0) + 1;
        const existingG = gScore.get(neighborKey);
        if (existingG !== undefined && tentativeG >= existingG) {
          continue;
        }

        cameFrom.set(neighborKey, currentKey);
        gScore.set(neighborKey, tentativeG);
        const fScore = tentativeG + this.heuristic(neighbor, goal);

        if (!openSet.has(neighborKey)) {
          open.push({ x: neighbor.x, y: neighbor.y, f: fScore });
          openSet.add(neighborKey);
        }
      }
    }

    return [];
  }

  private reconstructPath(cameFrom: Map<string, string>, goalKey: string) {
    const path: Array<{ x: number; y: number }> = [];
    let currentKey: string | undefined = goalKey;

    while (currentKey) {
      const [x, y] = currentKey.split(",").map(Number);
      path.push({ x, y });
      currentKey = cameFrom.get(currentKey);
    }

    return path.reverse();
  }

  private heuristic(a: { x: number; y: number }, b: { x: number; y: number }) {
    return Math.abs(a.x - b.x) + Math.abs(a.y - b.y);
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
