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

Would you like me to expand on the **"Server Architecture"** section regarding how to handle the **Game Loop** in Go?
