For a solo/small team project like this, you don't need a 50-page corporate binder. You need **"Living Documents"** that keep you focused.

I recommend creating exactly **three documents** in your repository (e.g., in a `/docs` folder).

Here are the templates tailored for your **RuneScape-like Web MMO**.

---

### Document 1: `GDD_LITE.md` (Game Design Document)

*Purpose: Prevents "Feature Creep." If it's not in this doc, you don't build it for the MVP.*

```markdown
# 📜 Game Design Document (MVP Scope)

## 1. The Core Loop
1.  **Gather:** Player clicks a resource node (Tree/Rock). Action takes time (ticks).
2.  **Process:** Player turns resource into product (Log -> Plank, Ore -> Bar).
3.  **Upgrade:** Player sells product for Gold -> Buys better Tool -> Gathers faster.
4.  **Combat:** Player uses gear to fight simple mobs for rare drops.

## 2. MVP Features (The "Must Haves")
### Skills (Vertical Progression)
* **Woodcutting:** Trees (Normal, Oak, Willow).
* **Mining:** Rocks (Copper, Tin, Iron).
* **Smithing:** Smelting bars, Anvil smithing (Bronze, Iron).
* **Combat (Melee):** Attack (Accuracy), Strength (Max Hit), Defense (Dmg Reduction).

### Systems
* **Inventory:** 28 slots. Items are drag-and-drop.
* **Bank:** Infinite storage (stackable items).
* **Movement:** A* Pathfinding on a grid. Click-to-move.
* **Chat:** Local chat (broadcast to nearby players).

## 3. The "Focus Layer" (Unique Mechanic)
* **Passive:** Auto-gather at 1x speed.
* **Active:** "Sparkles" appear on resources. Clicking them grants 2x yield + "Focus Token".
* **Horizontal:** Tokens are used to buy "Skill Tomes" (unlock special abilities).

## 4. World Map (Scope)
* **Town (Spawn):** 1 Bank, 1 General Store, 1 Anvil/Furnace.
* **Forest:** 50x50 tiles. Goblins (Lvl 2), Trees.
* **Mine:** 30x30 tiles. Rocks, Scorpions (Lvl 5).

```

---

### Document 2: `TECH_SPEC.md` (Architecture & Contracts)

*Purpose: Solves hard technical problems before you write code.*

```markdown
# ⚙️ Technical Specification

## 1. Stack
* **Frontend:** Next.js (App Router), Phaser 3 (Canvas), Zustand (State), Tailwind.
* **Backend:** Golang (Game Loop), WebSocket (Gorilla/Nhooyr).
* **Database:** PostgreSQL (Inventory/Skills), Redis (Chat/Session/Hot Coords).
* **Protocol:** JSON for MVP (switch to Protobuf later).

## 2. Server Architecture (ECS/Data-Oriented)
* **Tick Rate:** 600ms (Server Heartbeat).
* **Map System:** Chunk-based (32x32 tiles).
* **Entities:**
    * Players (Dynamic, Persisted).
    * NPCs (Dynamic, AI-driven).
    * StaticObjects (Trees, Rocks - State: Available/Depleted).

## 3. Database Schema (Postgres)
```sql
-- See 'schema.sql' for full details
-- Key Tables: users, characters, inventory_items, skill_xp

```

## 4. WebSocket Packets (The Contract)

### Client -> Server

* `MOVE_INTENT`: { x: 10, y: 20 }
* `INTERACT`: { targetId: 505, action: "chop" }
* `CHAT_MSG`: { text: "selling lobbies" }

### Server -> Client (Broadcast)

* `STATE_UPDATE`: { tick: 105, players: [{id: 1, x: 10, y: 20, anim: "walk"}] }
* `INVENTORY_UPDATE`: { slot: 0, itemId: 55, amount: 2 }

```

---

### Document 3: `ROADMAP.md` (The Plan)
*Purpose: Keeps you motivated by breaking the mountain into stairs.*

```markdown
# 📅 Development Roadmap

## Phase 1: The "Walking Simulator" (Weeks 1-2)
- [ ] Setup Go Server with WebSocket Loop.
- [ ] Setup Next.js + Phaser Bridge.
- [ ] Render a Tiled Map (exported from Tiled).
- [ ] Implement A* Pathfinding (Server-side validation).
- [ ] **Goal:** 2 Browser windows can see each other move.

## Phase 2: The "Lumberjack" (Weeks 3-4)
- [ ] Database persistence (Login/Save position).
- [ ] Object Interaction (Click Tree -> Tree turns to stump).
- [ ] Inventory UI (Zustand + Retro CSS).
- [ ] Resource adding (Server adds Log to DB -> UI updates).
- [ ] **Goal:** I can log in, chop a full inventory, and it saves.

## Phase 3: The "Economy" (Weeks 5-6)
- [ ] Bank Interface (Drag/Drop from Inventory).
- [ ] Dropping items on the ground.
- [ ] Simple NPC Shop (Sell logs for Gold).
- [ ] **Goal:** The core loop (Work -> Money) is complete.

## Phase 4: The "Fighter" (Weeks 7-8)
- [ ] Combat Stats (HP, Attack, Defense).
- [ ] Simple Mob AI (Chase player if close).
- [ ] Combat Loop (Tick-based damage calculation).
- [ ] Loot Tables.
- [ ] **Goal:** I can kill a Goblin and pick up its bones.

```

### Recommendation

Start by creating these three files in your repo today. It forces your brain to visualize the *finish line* of the MVP, rather than just the code you are writing right now.

# 🏗️ Technical Architecture & Directory Structure

## 1. High-Level Overview

**RuneWeb** (Project Name TBD) follows a **Monorepo** structure.

* **Backend:** A **Golang** server using a Hybrid Architecture (DDD for persistence, Data-Oriented/ECS for the Game Loop).
* **Frontend:** A **Next.js** application where **React** handles the UI/HUD and **Phaser 3** handles the game world rendering.
* **Communication:** WebSockets (Binary/Protobuf preferred) for real-time game state; REST/RPC for account management.

---

## 2. Directory Structure

```text
/project-root
├── /cmd                    # Application Entry Points
│   └── /server             # Main Game Server executable (go run ./cmd/server)
│
├── /internal               # Private Backend Code
│   ├── /domain             # [Slow Logic] DDD Layers (Transactional/Persistence)
│   │   ├── /account        # Auth, User Profiles
│   │   ├── /inventory      # Item logic, Trading, Banking
│   │   └── /market         # Grand Exchange logic
│   │
│   ├── /game               # [Fast Logic] The Real-Time Engine (Data-Oriented)
│   │   ├── /engine
│   │   │   ├── loop.go     # The 600ms Ticker
│   │   │   └── world.go    # Map Manager & Chunk Loading
│   │   ├── /systems        # Movement, Combat, Regeneration (ECS-like)
│   │   └── /spatial        # Spatial Hashing (Find players near X,Y)
│   │
│   ├── /network            # Connectivity
│   │   ├── /websocket      # Client connection manager
│   │   └── /packets        # Serializers (Proto/JSON)
│   │
│   └── /infrastructure     # Implementation details
│       ├── /postgres       # SQL Queries (sqlc generated)
│       └── /redis          # Cache & Hot Data
│
├── /pkg                    # Shared Code (Utilities, Math, Pathfinding)
├── /proto                  # Protocol Buffer Definitions (.proto files)
├── /migrations             # SQL Database Migrations
│
└── /web-client             # The Frontend Application (Next.js)
    ├── /public             # Static Assets (Sprites, Sounds, JSON Maps)
    ├── /src
    │   ├── /app            # Next.js App Router Pages
    │   │   ├── /play       # The Game Page (CSR - Client Side Rendered)
    │   │   ├── /market     # The Grand Exchange Portal (SSR - Server Side)
    │   │   └── layout.tsx  # Global Pixel Font Setup
    │   │
    │   ├── /components
    │   │   ├── /game       # The Phaser Wrapper (<PhaserGame />)
    │   │   └── /ui         # React HUD Components (Inventory, Chat, Stats)
    │   │       └── /retro  # Custom 9-Slice CSS Components
    │   │
    │   ├── /game-engine    # PURE Phaser Logic (No React code)
    │   │   ├── game.ts     # Phaser Config
    │   │   ├── /scenes     # BootScene, MainScene
    │   │   ├── /systems    # InputManager, NetworkManager
    │   │   └── /entities   # PlayerSprite, MobSprite
    │   │
    │   └── /store          # The "Bridge" (Zustand)
    │       └── gameStore.ts

```

---

## 3. Backend Architecture: The "Hybrid" Model

We avoid pure DDD for the game loop to prevent Garbage Collection pauses. We split the logic into two distinct lifecycles.

### A. The "Outer Loop" (Domain-Driven Design)

Used for low-frequency, high-integrity actions.

* **Use Cases:** Login, Trading, Banking, Quest Completion.
* **Pattern:** Controller -> Service -> Repository -> Database.
* **Storage:** PostgreSQL (Strict ACID compliance).

### B. The "Inner Loop" (Game Engine)

Used for high-frequency, performance-critical actions.

* **Use Cases:** Movement, Collision, Combat Calculations, Aggro range.
* **Pattern:** The **Game Loop** (Heartbeat).
1. **Process Inputs:** Collect all "Move Intents" from the buffer.
2. **Update State:** Resolve movement (A*), apply damage, tick timers.
3. **Broadcast:** Send the new World State snapshot to relevant clients.


* **Storage:** In-Memory structs (Map chunks), Redis (Ephemeral state).

---

## 4. Frontend Architecture: The "Bridge" Pattern

We do not render the game world in React. We do not render the UI in Phaser. We bridge them.

### A. The Separation of Concerns

| Feature | Technology | Rendering | State Source |
| --- | --- | --- | --- |
| **Game World** | **Phaser 3** | `<canvas>` (WebGL) | WebSocket Packets |
| **UI (HUD)** | **React** | HTML DOM (Floating) | Zustand Store |
| **Meta Pages** | **Next.js** | SSR HTML | Server Database |

### B. The "Bridge" (Zustand)

Since Phaser runs outside the React Component Tree, it cannot use `useContext`. We use **Zustand** as a globally accessible store.

1. **Server** sends packet: `{"type": "HP_UPDATE", "val": 45}`
2. **Phaser** receives packet -> calls `gameStoreApi.getState().setHp(45)`
3. **React Component** (`<HealthBar />`) subscribes to `useStore(state => state.hp)`
4. **UI Updates** instantly without re-rendering the Canvas.

---

## 5. UI Architecture: "Retro-Modern"

### A. Styling Strategy

* **No 8-bit Libraries:** We avoid pre-made 8-bit UI libraries due to poor readability and layout constraints.
* **9-Slice Scaling:** We use standard HTML divs with CSS `border-image` to create scalable, pixel-perfect retro windows.
* **Tailwind CSS:** Used for layout (Flexbox/Grid) and positioning.

### B. The "Floating HUD"

* **Container:** `pointer-events-none` (Allows clicks to pass through to the game world).
* **Windows:** `pointer-events-auto` (Captures clicks for Inventory/Chat).
* **Strategy:** UI elements (Minimap, Chat, Inventory) float over the full-screen canvas.

---

## 6. Database Strategy

### A. PostgreSQL (The Source of Truth)

* **Users:** Auth data.
* **Characters:** `x, y, region_id`, appearance.
* **Inventory:** Composite key (`char_id` + `slot_id`) to ensure data integrity.
* **Skills:** Experience points table.

### B. Redis (The Hot Layer)

* **Grand Exchange Order Book:** Uses `Sorted Sets` for O(1) order matching.
* **Chat Buffers:** Short-term history for when a player logs in.
* **Session Data:** "Who is online and on which Map Chunk?"

---

## 7. Networking (The Protocol)

* **Transport:** WebSockets (TCP).
* **Tick Rate:** 600ms (Server Logic) / 60fps (Client Interpolation).
* **Movement Logic:**
1. **Client:** Player clicks -> Calculate A* path locally -> Start moving (Prediction) -> Send `Target(x,y)` to server.
2. **Server:** Validate path (Anti-cheat) -> Tick movement -> Broadcast `PlayerPos(x,y)`.
3. **Reconciliation:** If Client and Server disagree (e.g., Lag/Hack), Client "snaps" to Server position.
