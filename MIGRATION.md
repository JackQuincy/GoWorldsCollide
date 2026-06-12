# Worlds Collide Go Migration

This file is the persistent handoff for continuing the Go rewrite in a new
session.

## Goal

Rewrite the Python Worlds Collide randomizer in Go while preserving observable
behavior. Prefer direct, test-backed ports over redesigns. Exact NumPy random
bitstream compatibility is not currently required, but seeded Go runs must be
deterministic.

## Source Of Truth

- Upstream repository: `https://github.com/JackQuincy/WorldsCollide`
- Source branch: `worlds-divided`
- Do not use upstream `main` as the behavioral source unless explicitly asked.
- A temporary clone may be made outside this repository for source inspection.

## Decisions

- Module path: `github.com/JackQuincy/GoWorldsCollide`
- Minimum Go version: Go 1.22
- Implementation packages belong under `internal/`.
- Randomness goes through `internal/random`.
- RNG implementation: `math/rand/v2` PCG.
- Seed text is SHA-256 hashed following upstream `rng.py`; the first 16 bytes
  initialize the Go PCG wrapper.
- Bit-for-bit parity with NumPy PCG64 is deferred.
- Tests use synthetic byte slices. Never commit ROMs or derived ROM files.
- Preserve upstream inclusive address ranges and allocation order.
- Use Python-generated golden values for instruction and binary parity tests.

## Completed

### Migration Foundation

Commit `b734416` (`Start Go migration foundation`)

- Added `AGENTS.md` and `go.mod`.
- Added the seeded PCG wrapper in `internal/random`.
- Ported weighted random selection, conditional shuffle, intersection, and
  truncated discrete distribution helpers.
- Ported ROM validation, expansion, reads, writes, endian helpers, and bit
  operations.
- Ported heap allocation and the complete upstream free-space map.

### Labels And Spaces

Commit `97d34a0` (`Port labels and memory spaces`)

- Ported all label pointer modes.
- Added forward label fixups and pointer offset handling.
- Added bank-aware reserve and allocate operations.
- Added overlap detection, nested writes, clear/copy operations, and SNES
  address conversion.
- Added the `memory.Instruction` interface used by instruction encoders.

### 65816 Assembly

Commit `9be5e3f` (`Port 65816 assembly instructions`)

- Ported all 20 addressing modes from `instruction/asm.py`.
- Ported every upstream instruction constructor and alias.
- Added operand masking and overflow validation.
- Added Python golden-vector tests.
- Added forward and backward branch tests through `memory.Space`.

### Event Instruction Foundation

Commit `62edc11` (`Port event instruction foundation`)

- Ported recursive event instruction argument flattening.
- Ported event branches with absolute addresses and symbolic 24-bit label
  pointers relative to `EVENT_CODE_START`.
- Ported shared map-load direction, music, transition, vehicle, and parent-map
  flag packing.
- Added Python golden-vector tests and forward label-resolution coverage.

### Entity Instructions

Commit `707641a` (`Port entity instructions`)

- Ported entity action termination and pause instructions.
- Ported movement and facing-direction opcode packing.
- Preserved the upstream distance-eight clamp and warning.
- Ported all five entity movement speeds.
- Added Python golden-vector and boundary tests for every constructor.

### Vehicle Instructions

Commit `099b129` (`Port vehicle instructions`)

- Ported vehicle position and action termination instructions.
- Ported set/clear event-bit instructions with upstream range validation.
- Ported conditional and unconditional event-code branches, including symbolic
  label resolution.
- Ported fade and non-fade map loading through the shared event encoder.
- Added Python golden-vector, event-bit boundary, and label-resolution tests.

### World, Battle Event, And Field Entity Instructions

- Ported the complete world-map instruction surface, including shared entity
  actions, castle commands, map loading, and event-bit branches.
- Ported battle-event targets, animation slots, dialogs, and script control.
- Replaced upstream battle-event import-time ROM mutation with an explicit
  `battleevent.InstallHandlers` operation.
- Ported the C1 add/remove-target relocation, checks-complete handler, and
  opcode-table updates.
- Ported the complete field entity action surface, including all animations,
  diagonal movement, entity constants, and label-based distance branches.
- Added Python golden vectors, label tests, animation-slot tests, and exact C1
  patch-output tests.

At this checkpoint, `go test ./...` and `go vet ./...` pass.

## Deferred Work

- `utils/flatten.py` is intentionally deferred. Python uses it on heterogeneous
  nested instruction values. `memory.Space.Write` currently provides the needed
  Go flattening behavior.
- Exact NumPy PCG64 output parity is deferred behind `internal/random`.
- ROM-backed integration tests requiring copyrighted input must not be added.

## Migration Roadmap

Follow this order unless dependencies discovered in the source require a small
adjustment.

1. Port the remaining instruction foundations:
   - `instruction/event.py`
   - `instruction/entity.py`
   - `instruction/vehicle.py`
   - `instruction/world.py`
   - `instruction/battle_event.py`
   - `instruction/field/`
2. Port bank-specific instruction and patch modules:
   - `instruction/c0.py`
   - `instruction/c1.py`
   - `instruction/c2.py`
   - `instruction/c3.py`
   - `instruction/c4.py`
   - `instruction/f0.py`
3. Port simple constants and binary data structures:
   - `constants/`
   - `data/structures.py`
   - small fixed-record data models
4. Port ROM data codecs by domain:
   - items, spells, espers, characters
   - enemies, formations, packs, scripts
   - maps, NPCs, exits, events, dialogs
5. Port isolated patches as early vertical slices:
   - begin with a small module from `bug_fixes/`
   - load a ROM, apply one patch, and write output
6. Port randomization domains:
   - chests, shops, commands, espers, enemies
   - objectives and rewards
   - events and world division logic
7. Port CLI and orchestration:
   - `args/`
   - `settings/`
   - seed generation
   - `cmd/wc`
8. Port presentation and supporting systems:
   - menus, graphics, music, metadata, API helpers, and logging
9. Add end-to-end parity checks using legal checksums, offsets, patch bytes, and
   deterministic output metadata.

## Next Task

Port the core `instruction/field/instructions.py` surface.

Suggested workflow:

1. Group constructors by script control, party, inventory, entity, display,
   audio, map, event-bit, event-word, battle, and timer behavior.
2. Implement pure encoders under `internal/instruction/field`.
3. Reuse `event.NewInstruction`, `event.NewBranch`, and `event.NewLoadMap`.
4. Keep dynamic constructor families explicit and test generated opcode/bit
   combinations against Python.
5. Port helper instruction sequences only after their primitive encoders exist.
6. Defer `field/custom.py` and `field/y_npc/` patch installers until their C0
   and bank-specific dependencies are available.
7. Run:

```sh
gofmt -w .
go test ./...
go vet ./...
```

## Session Checklist

At the start of a session:

1. Read `AGENTS.md` and this file.
2. Run `git status --short`.
3. Inspect the relevant file from upstream `worlds-divided`.
4. Run `go test ./...` before editing if practical.

Before handing work back:

1. Run `gofmt` on changed Go files.
2. Run focused package tests.
3. Run `go test ./...`.
4. Run `go vet ./...`.
5. Update this file when a roadmap phase or important decision changes.
6. Keep commits focused and record completed milestones here.
