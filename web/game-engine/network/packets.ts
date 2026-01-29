export const PacketMoveIntent = "MOVE_INTENT";
export const PacketStateSnapshot = "STATE_SNAPSHOT";
export const PacketStateDelta = "STATE_DELTA";
export const PacketWelcome = "WELCOME";
export const POSITION_SCALE = 100;

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

export type StateSnapshot = {
  tick: number;
  players: PlayerState[];
};

export type StateDelta = {
  tick: number;
  players: PlayerState[];
  removed: string[];
};

export type Welcome = {
  id: string;
};
