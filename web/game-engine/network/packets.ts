export const PacketMoveIntent = "MOVE_INTENT";
export const PacketStateUpdate = "STATE_UPDATE";
export const PacketWelcome = "WELCOME";

export type Packet<T = unknown> = {
  type: string;
  payload: T;
};

export type MoveIntent = {
  x: number;
  y: number;
};

export type PlayerState = {
  id: string;
  x: number;
  y: number;
};

export type StateUpdate = {
  tick: number;
  players: PlayerState[];
};

export type Welcome = {
  id: string;
};
