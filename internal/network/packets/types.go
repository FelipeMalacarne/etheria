package packets

import "encoding/json"

const (
	PacketMoveIntent  = "MOVE_INTENT"
	PacketStateSnapshot = "STATE_SNAPSHOT"
	PacketStateDelta    = "STATE_DELTA"
	PacketWelcome     = "WELCOME"
)

type Packet struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type MoveIntent struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type PlayerState struct {
	ID string `json:"id"`
	X  int `json:"x"`
	Y  int `json:"y"`
}

type StateSnapshot struct {
	Tick    int64         `json:"tick"`
	Players []PlayerState `json:"players"`
}

type StateDelta struct {
	Tick    int64         `json:"tick"`
	Players []PlayerState `json:"players"`
	Removed []string      `json:"removed"`
}

type Welcome struct {
	ID string `json:"id"`
}

func NewPacket(packetType string, payload any) (Packet, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Packet{}, err
	}

	return Packet{
		Type:    packetType,
		Payload: data,
	}, nil
}
