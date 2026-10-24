# Thirty Nights tilemap editor

Browser-based editor for the game's tile maps. It imports a tileset image, paints
layers on a canvas, validates the result, and downloads JSON that loads through
`world.LoadMap`.

No backend is involved. The map never leaves the browser.

## Run it

```sh
cd TileMapEditor
bun install
bun run dev
```

Open the URL printed by Vite. Other useful commands:

```sh
bun run check   # svelte-check + TypeScript
bun run test    # vitest
bun run lint    # prettier + eslint
bun run build
```

## Workflow

1. Set the map ID, name, dimensions, tile size, and spawn point on the left.
2. Import a PNG or JPEG tileset in the right panel.
3. Set source tile width/height, spacing, and margin, then press **Apply**.
4. Click a tile to make it the paint tile. Edit its ID, colour, and solid flag
   in the **Tile inspector** tab.
5. Paint on the workspace. Drag paints a continuous line; use **Erase** to set
   cells back to the empty string.
6. Add, rename, reorder, hide, and delete layers in the layer panel.
7. Fix anything listed under **Validate & export**, then export the JSON.

Mouse wheel zooms around the pointer. Middle-drag, right-drag, or space-drag
pans. Double-clicking the workspace fits the map to the viewport.

## What the game expects

Export mirrors the Go structs in `src/world/data.go`:

- `format_version`, `id`, `name`, `tile_size`, `width`, `height`
- `tiles`: a map of tile ID to `{ solid, visual: { color } }`
- `layers`: `{ id, type, z_index, collision, grid }`, with row-major
  `grid[y][x]` and `""` for empty cells
- `rooms`: `{ id, name, bounds }`
- `objects`: `{ id, type, layer, x, y, properties }`
- `spawn`: `{ x, y }`

The game draws tiles from `visual.color`. The imported tileset is an editor
reference: the **ART** toggle overlays it on the workspace, while exported tiles
keep the sampled colour. Source tile dimensions and the map's logical
`tile_size` are separate values.

Validation is stricter than serialization on purpose. It follows the Go
loader's rules for map metadata, tiles, layers, rooms, and supported object
types, then adds editor safeguards: spawn bounds, object-layer references, and
practical size limits. Export is disabled until there are no errors.

## Code layout

- `src/lib/map-model`: game map types and blank-map construction
- `src/lib/map-operations`: cell, layer, tile, and resize mutations
- `src/lib/map-validation`: loader-compatible validation
- `src/lib/map-serializer`: exact export shape and filename handling
- `src/lib/tileset`: image slicing and source geometry
- `src/lib/renderer`: viewport math and canvas drawing
- `src/lib/editor-state.svelte.ts`: editor state, tileset import, painting, export
- `src/lib/ui`: Svelte panels, canvas input, and controls
- `src/lib/file-io.ts`: image decoding, colour sampling, JSON download

The canvas is only a rendering surface. The in-memory map is authoritative, and
zoom, pan, hover, selection, and layer visibility stay in editor state so they
cannot leak into the export.

## Current limits

Undo/redo, project save/reopen, map import, object placement, and room editing
are not implemented. The map model and serializer already carry rooms and
objects so existing maps round-trip once an import path is added.
