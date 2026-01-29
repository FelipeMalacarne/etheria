import {
  Packet,
  PacketMoveIntent,
  PacketStateUpdate,
  PacketWelcome,
  StateUpdate,
  Welcome,
} from "./packets";

type NetworkHandlers = {
  onStateUpdate?: (update: StateUpdate) => void;
  onWelcome?: (welcome: Welcome) => void;
  onConnectionChange?: (connected: boolean) => void;
};

export class NetworkClient {
  private socket: WebSocket | null = null;
  private handlers: NetworkHandlers;

  constructor(handlers: NetworkHandlers) {
    this.handlers = handlers;
  }

  connect(url: string) {
    if (this.socket) {
      return;
    }

    const socket = new WebSocket(url);
    this.socket = socket;

    socket.addEventListener("open", () => {
      this.handlers.onConnectionChange?.(true);
    });

    socket.addEventListener("close", () => {
      this.handlers.onConnectionChange?.(false);
      this.socket = null;
    });

    socket.addEventListener("error", () => {
      this.handlers.onConnectionChange?.(false);
    });

    socket.addEventListener("message", (event) => {
      if (typeof event.data !== "string") {
        return;
      }

      let packet: Packet;
      try {
        packet = JSON.parse(event.data) as Packet;
      } catch {
        return;
      }

      switch (packet.type) {
        case PacketStateUpdate:
          this.handlers.onStateUpdate?.(packet.payload as StateUpdate);
          break;
        case PacketWelcome:
          this.handlers.onWelcome?.(packet.payload as Welcome);
          break;
        default:
          break;
      }
    });
  }

  disconnect() {
    if (!this.socket) {
      return;
    }

    this.socket.close();
    this.socket = null;
  }

  sendMoveIntent(x: number, y: number) {
    this.send(PacketMoveIntent, { x, y });
  }

  private send<T>(type: string, payload: T) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return;
    }

    const packet: Packet<T> = { type, payload };
    this.socket.send(JSON.stringify(packet));
  }
}
