# AGENTS.md

Guidance for agents working on the Go rewrite of
[JackQuincy/WorldsCollide](https://github.com/JackQuincy/WorldsCollide).

The current goal is a careful port, not a redesign. Preserve behavior from the
Python project unless a change is explicitly requested, and document intentional
differences near the code that introduces them.

## Project Shape

The upstream Python repository is organized around these domains:

- `wc.py`, `seed.py`, `valid_rom_file.py`, `version.py`: command entrypoint,
  seed generation, ROM validation, and version metadata.
- `args/`: command-line flag definitions and option parsing.
- `settings/`: game settings derived from flags.
- `data/`: ROM-backed data models and readers/writers for characters, items,
  espers, enemies, maps, shops, dialogs, spells, rages, lores, and related
  tables.
- `event/`: event-specific randomization and patch logic.
- `instruction/`: assembly/event/field instruction encoders.
- `memory/`: ROM byte access, address allocation, labels, free space, and patch
  placement.
- `battle/`: battle mechanics patches, scaling, checks, and multipliers.
- `objectives/`: objective conditions, results, and completion checks.
- `graphics/`: sprites, palettes, portraits, title graphics, and conversion
  tools.
- `menus/`: menu and in-game display modifications.
- `bug_fixes/`: isolated patches for known game bugs.
- `constants/`: canonical IDs, names, flags, and domain constants.
- `metadata/`, `api/`, `music/`, `utils/`, `log/`: supporting utilities,
  metadata generation, web-facing helpers, music helpers, and logging.

Use this Go layout unless a better structure emerges during the port:

```text
.
+-- AGENTS.md
+-- LICENSE
+-- README.md
+-- go.mod
+-- cmd/
|   +-- wc/                  # CLI equivalent of wc.py
|   +-- wcflagmeta/          # metadata generator, if retained
+-- internal/
|   +-- app/                 # orchestration for a randomizer run
|   +-- args/                # flag parsing and validation
|   +-- battle/              # battle patch logic
|   +-- bugfix/              # isolated bug-fix patches
|   +-- constants/           # generated/static IDs and lookup tables
|   +-- data/                # ROM data models and codecs
|   +-- event/               # event randomization and patch logic
|   +-- graphics/            # sprites, palettes, portraits, conversions
|   +-- instruction/         # instruction encoders and assemblers
|   +-- log/                 # logging helpers
|   +-- memory/              # ROM, free-space, labels, and writes
|   +-- menu/                # menu patches
|   +-- metadata/            # flag/objective metadata generation
|   +-- music/               # song helpers
|   +-- objective/           # objectives, conditions, and results
|   +-- random/              # seeded RNG and weighted choices
|   +-- settings/            # normalized settings from parsed args
+-- pkg/
|   +-- worlds/              # only stable public APIs, if any are needed
+-- testdata/                # tiny fixtures only; no copyrighted ROMs
+-- tools/                   # developer-only generators or converters
```

Prefer `internal/` for implementation packages. Add `pkg/` only for APIs that
external callers are expected to import.

## Porting Practices

- Port behavior in small vertical slices: parse options, load ROM data, apply
  one patch family, write output, then test that slice.
- Keep package names short, singular, and lower-case: `memory`, `event`,
  `objective`, `bugfix`.
- Keep Python names visible in comments only when useful for traceability, for
  example when a Go function is a direct port of a specific Python function.
- Use explicit structs for ROM records. Avoid map-heavy representations when a
  fixed binary layout is known.
- Put binary offsets, bank constants, IDs, and bit masks in named constants.
  Do not leave magic numbers in patch logic.
- Treat ROM mutations as byte-level I/O with clear ownership. A function should
  either compute data or write to the ROM, not quietly do both unless its name
  says so.
- Use `io.Reader`, `io.Writer`, `[]byte`, and small interfaces at package
  boundaries. Avoid global mutable state.
- All randomness must flow from an injected seeded RNG. Do not call package-level
  random functions from ported logic. Use `internal/random`, which wraps
  `math/rand/v2` with a PCG source.
- Return errors with context using `fmt.Errorf("read enemy %d: %w", id, err)`.
  Do not panic for user input, corrupt ROMs, or invalid flags.
- Keep generated metadata and generated constants reproducible. Generators belong
  in `tools/` or `cmd/`, and generated files must say how to regenerate them.

## Go Style

- Run `gofmt` on every Go file before finishing a change.
- Run `go test ./...` before handing work back when the module exists.
- Prefer table-driven tests for parsers, codecs, RNG choices, and patch
  builders.
- Keep functions focused. If a direct Python port is large, first preserve
  behavior, then split along meaningful domain boundaries once tests exist.
- Use standard library packages first. Add dependencies only when they remove
  meaningful complexity.
- Use `context.Context` only for operations that can be cancelled or have
  deadlines, such as CLI orchestration or external I/O.
- Keep CLI output deterministic where possible so tests can assert it.

## Testing Expectations

Every non-trivial package should have unit tests next to the code:

```text
internal/memory/rom.go
internal/memory/rom_test.go
internal/data/items.go
internal/data/items_test.go
```

Add tests for:

- Binary read/write codecs, including boundary offsets and malformed input.
- Flag parsing and validation, especially defaults and conflicting options.
- Seed generation and deterministic RNG behavior.
- Weighted random selection and shuffling behavior with fixed seeds.
- Patch builders, by asserting exact byte output for small fixtures.
- Objective condition/result evaluation.
- Error cases for short reads, invalid IDs, invalid flags, and exhausted memory
  space.

Use `testdata/` for small synthetic fixtures. Do not commit commercial ROMs,
derived ROMs, save files, or large copyrighted assets. If a test needs ROM-like
data, build the minimal byte slice in the test or place a tiny artificial binary
fixture in `testdata/`.

When porting from Python, keep a parity test whenever possible:

- Capture expected values from the original implementation for tiny, legal
  inputs.
- Assert deterministic output for the same seed and flags.
- Prefer exact byte comparisons for low-level encoders and patch output.

## Commands

Use these commands once the Go module exists:

```sh
go test ./...
go test -race ./...
go vet ./...
gofmt -w .
```

For focused work, run the narrow package test first, then the full suite:

```sh
go test ./internal/memory
go test ./...
```

## Git And Repository Hygiene

- Keep commits focused on one domain or vertical slice.
- Do not commit generated output unless it is required by the build or CLI.
- Do not commit ROM files or patched ROM outputs.
- Keep `README.md` user-focused. Put agent and contributor workflow details in
  this file or future docs under `docs/`.
- If upstream behavior is unclear, inspect the Python source and add a test that
  locks in the discovered behavior before porting.
