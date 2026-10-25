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

1. Set the map ID, name, dimensions, tile size, and spawn point on the left,
   or press **Import** to load an exported map JSON file.
2. Drop a PNG or JPEG onto the tileset drop zone in the right panel, or click it
   to browse.
3. Set source tile width/height, spacing, and margins, then press **Apply**.
4. Click a tile to make it the paint tile. Edit its ID, colour, and solid flag
   in the **Tile inspector** tab.
5. Paint on the workspace. Drag paints a continuous line; use **Erase** to set
   cells back to the empty string and remove objects under the brush.
6. Add, rename, drag to reorder, hide, and delete layers in the layer panel.
7. In the **Objects** tab, create object types and click a type to place it.
   Objects go into the map's `objects` list, not into layer grids. Drag places a
   line of objects; right-click or **Erase** removes them.
8. Fix anything listed under **Validate & export**, then export the JSON.

Mouse wheel zooms around the pointer. Middle-drag or space-drag pans.
Right-click applies the alternate tool: paint erases, erase paints.
Ctrl+Z (or Cmd+Z) undoes the last edit; Ctrl+Y (or Cmd+Y) redoes it. A drag counts as one step.
Double-clicking the workspace fits the map to the viewport.

Single-key shortcuts switch tools: **B** paint, **E** erase, **O** object,
**S** spawn, **G** toggles the grid, **+**/**-** zoom, **1** returns to 100% zoom.
The menu bar mirrors these actions under File, Edit, View, and Tools, including
Ctrl+N (new map), Ctrl+O (import), and Ctrl+E (export).

Both sidebars resize by dragging the divider between them and the canvas, and
the Layers panel resizes by dragging the divider above it. Widths and the layer
height persist per browser; double-clicking a divider restores its default.

The current map is autosaved to browser storage at most once every five
seconds, and pending edits are flushed when the page is hidden or unloaded, so
reloading or reopening the editor restores your work. Ctrl+S saves immediately.
Change the autosave delay under **Edit → Settings…**.

The same dialog holds the map JSON schema that drives export. It is a plain
JSON Schema document with a small code editor, an upload button, and a reset to
the built-in schema. Export follows the schema's `properties` order: known map
fields use the edited values, and extra properties are written with their
`default`. Fields the UI depends on (`format_version`, `id`, `name`,
`tile_size`, `width`, `height`, `tiles`, `layers`, `objects`, `spawn`) must
stay in both `properties` and `required`, and any other required property needs
a `default`. The schema is stored in `localStorage`; there is no full JSON
Schema type validation.

## What the game expects

Export mirrors the Go structs in `src/world/data.go`:

- `format_version`, `id`, `name`, `tile_size`, `width`, `height`
- `tiles`: a map of tile ID to `{ solid, visual: { color } }`
- `layers`: `{ id, type, z_index, collision, grid }`, with row-major
  `grid[y][x]` and `""` for empty cells
- `objects`: `{ id, type, layer, x, y, properties }`
- `spawn`: `{ x, y }`

Rooms are no longer emitted or read. The Go loader ignores a missing `rooms`
field, and leftover `room_id` properties on imported objects round-trip
unchanged. Object types are an editor palette only; they are not written to the
JSON.

The game draws tiles from `visual.color`. The imported tileset is an editor
reference: the **ART** toggle overlays it on the workspace, while exported tiles
keep the sampled colour. Source tile dimensions and the map's logical
`tile_size` are separate values.

Validation is stricter than serialization on purpose. It follows the Go
loader's rules for map metadata, tiles, layers, and object-layer references,
then adds editor safeguards: spawn bounds, object bounds, and practical size
limits. Object types are not checked against the game loader; the editor lets
any type through. Export is disabled while there are errors.

## Code layout

- `src/lib/map-model`: game map types and blank-map construction
- `src/lib/map-operations`: cell, layer, tile, object, and resize mutations
- `src/lib/map-validation`: loader-compatible validation
- `src/lib/map-serializer`: exact export shape and filename handling
- `src/lib/map-schema`: default export schema, schema checks, schema-driven export
- `src/lib/map-importer`: JSON parsing and shape normalization for map import
- `src/lib/tileset`: image slicing and source geometry
- `src/lib/renderer`: viewport math and canvas drawing
- `src/lib/editor-state.svelte.ts`: editor state, tileset import, painting, export
- `src/lib/ui`: Svelte panels, canvas input, and controls
- `src/lib/file-io.ts`: image decoding, colour sampling, JSON download

The canvas is only a rendering surface. The in-memory map is authoritative, and
zoom, pan, hover, selection, and layer visibility stay in editor state so they
cannot leak into the export.

## Current limits

Undo and redo cover map edits, including layer, tile, object, and slicing changes; project save/reopen is not implemented. Map import reads the
exported game schema, ignores retired fields such as `rooms`, and rebuilds the
object palette from the imported objects. Object palette types still live in
editor state, so they reset when a new map is created or another map is
imported, and they are not saved back into the exported JSON.
