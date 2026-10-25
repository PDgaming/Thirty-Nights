# Tilemap Editor — Product Specification

## 1. Overview

### 1.1 Purpose

A browser-based tilemap editor for designing levels for the game and exporting them as JSON files that conform to the game's existing map data schema.

The editor should provide a desktop-oriented, graphics-editor-style interface inspired by Photoshop. Users should be able to import a tileset, configure map dimensions, paint tiles onto a grid, manage layers, edit tile properties, and export a map ready to be loaded by the game.

### 1.2 Primary goals

- Provide an intuitive visual environment for constructing tile-based levels.
- Support importing image-based tilesets and selecting individual tiles.
- Support multiple map layers, with independent visibility controls.
- Allow users to configure map and tile metadata.
- Export valid, game-compatible JSON files.
- Keep the interface responsive while editing maps with many cells.

### 1.3 Target users

Developers and level designers working on the game.

### 1.4 Platform

A desktop-oriented web application that runs in a modern browser. The initial release does not require user accounts, cloud storage, or a backend.

## 2. Core workflow

1. Create a new map and configure its dimensions.
2. Upload a tileset image.
3. Configure the tileset's source tile dimensions and slicing settings.
4. Select a tile from the tileset.
5. Configure the selected tile's properties.
6. Paint tiles onto the map using the workspace.
7. Create, select, hide, and show map layers.
8. Configure the spawn point and other supported map metadata.
9. Validate the map and export it as a JSON file.

The application should make it clear which tile, layer, and editing settings are currently active.

## 3. User interface and layout

### 3.1 General layout

The application should use a three-panel layout inspired by Photoshop and similar desktop graphics editors.

- **Left panel:** Map configuration and export tools.
- **Center workspace:** Interactive map editor.
- **Right panel:** Tileset browser, selected-tile properties, and layer management.

The center workspace should occupy the majority of the available screen space. The side panels should have independently scrollable content where necessary.

The interface should use a dark, neutral theme, compact controls, subtle borders, and clear visual indicators for selected tools, tiles, and layers.

### 3.2 Left panel — Map settings

The left panel contains the map's global configuration.

**Map properties**

- Map ID.
- Map name.
- Map width in grid cells.
- Map height in grid cells.
- Map tile size, as required by the game's map schema.
- Spawn point coordinates.
- Any additional metadata required by the game.

**Spawn point**

Users should be able to configure the player's starting position through numeric coordinate inputs. If practical, they should also be able to select the spawn position directly in the workspace.

The editor must validate that the spawn point lies within the map boundaries.

**Export**

An export button should validate the current map and download a JSON file.

- The user should be able to provide or edit the exported filename.
- Invalid maps should produce actionable validation errors.
- Export must not silently omit required map fields or metadata.
- The exported JSON must conform to the game's expected data structure.

### 3.3 Center workspace — Map editor

The workspace displays the map as a two-dimensional grid of cells.

**Grid and coordinates**

- The grid has `map_width` columns and `map_height` rows.
- Each cell corresponds to one position in the game's tile grid.
- The rendered size of a cell may change with the editor's zoom level, but its logical map coordinates must remain unchanged.
- The workspace should display grid boundaries when useful, particularly at higher zoom levels.
- The map should remain correctly aligned during zooming and panning.

**Tile placement**

- Selecting a tile in the tileset browser makes it the active painting tile.
- Hovering over a valid cell displays a translucent preview of the selected tile.
- Clicking a cell places the selected tile on the active layer.
- Clicking and dragging should support continuous painting across cells.
- The user should be able to erase a cell or restore it to the layer's empty state.
- Painting should respect layer visibility, selection, and any relevant editing restrictions.
- **Ctrl+Z** (or **Cmd+Z**) should undo the last edit and **Ctrl+Y** (or **Cmd+Y**) should redo it. A paint or object drag counts as one step.

**Zoom and navigation**

- The mouse wheel should zoom the workspace in and out.
- Zooming should ideally remain centered around the pointer position.
- The workspace should support panning when the map extends beyond the visible viewport.
- Zooming must not modify map dimensions or tile coordinates.

**Rendering**

The workspace must render the map, grid, active-layer editing preview, and tile-placement preview independently enough that frequent mouse movements do not require rebuilding the entire application interface.

Only the visible portion of a large map should need to be drawn.

### 3.4 Right panel — Tileset and tile properties

The right panel should contain two primary tabs.

#### Tab A: Tile inspector

Displays information about the selected tileset tile.

**Tile preview**

- Show a larger preview of the selected tile.
- Display the tile's ID and source coordinates within the tileset.
- Clearly indicate which tile is currently selected.

**Tile properties**

- `tile_id`: Unique identifier used by the game to reference the tile.
- `solid`: Whether the tile blocks movement, according to the game's collision rules.
- Other tile metadata supported by the game's schema.

The inspector should allow users to edit these properties. Changes must be reflected in the map's exported tile definitions.

The editor should prevent duplicate tile IDs and references to undefined tiles.

#### Tab B: Tileset browser

Displays the complete imported tileset image and its individual tiles.

**Tileset import**

- Allow users to upload an image from their local device.
- Accept common browser-supported image formats, including PNG and JPEG.
- Display the imported image in the browser without requiring a server upload.
- Report unsupported files or decoding failures clearly.

**Tileset slicing**

Users should be able to configure:

- Source tile width.
- Source tile height.
- Optional spacing between tiles.
- Optional margin around the tileset image, configurable separately for the x and y axes.

The editor should calculate the number of available tiles from the image dimensions and slicing configuration.

Tiles should be displayed in a selectable grid. Selecting a tile should update the active tile and the tile inspector.

The source tile dimensions must be distinguished from the map's logical tile size. The editor must handle these values according to the game's rendering rules rather than assuming they are interchangeable.

### 3.5 Right panel — Layer management

Layer management should be available in a dedicated section below the tile-related tabs.

**Layer operations**

- Display all map layers in their rendering order.
- Select a layer to make it the active editing layer.
- Toggle a layer's visibility.
- Create a new layer.
- Rename a layer.
- Delete a layer, with confirmation when it contains placed tiles.
- Reorder layers when supported by the game's rendering model.

The selected layer must be visually distinguishable from other layers.

Hiding a layer should affect workspace rendering only. It must not delete its contents or remove it from the exported JSON.

The editor must preserve the layer ordering expected by the game.

## 4. Map data and state management

The application should maintain a structured in-memory representation of the map rather than treating the rendered canvas as the source of truth.

The map model should contain the information needed to produce the game's JSON format, including:

- Map metadata and dimensions.
- Tile definitions and their properties.
- Layer definitions and grid contents.
- Spawn point information.
- Rooms and objects, where supported by the game's schema.

Each grid cell should reference a valid tile ID or represent an empty cell using the representation expected by the game.

The editor must keep visual state separate from serialized game data. Zoom, pan position, hover previews, selected tiles, and hidden-layer display settings should not accidentally become part of the exported map.

Changes made through the interface should update the in-memory map model. Rendering should reflect that model.

## 5. JSON export and validation

### 5.1 Export requirements

The export process must serialize the in-memory map into the exact structure expected by the game's Go map loader.

The current game schema includes the following top-level fields:

- `format_version`
- `id`
- `name`
- `tile_size`
- `width`
- `height`
- `tiles`
- `layers`
- `rooms`
- `objects`

The editor must preserve the exact field names, data types, nesting, and conventions required by the game.

The editor's internal names, such as `map_width` and `map_height`, may be mapped to the schema's `width` and `height` fields during export.

The exact placement and representation of the spawn point must follow the game's existing schema. It must not be introduced as a new top-level field unless the game explicitly supports that field.

### 5.2 Validation rules

Before export, validate at least the following:

- Map dimensions are positive integers within supported limits.
- The map has the required metadata and format version.
- Tile definitions have valid, unique IDs.
- Every non-empty grid reference points to an existing tile definition.
- Every layer has valid dimensions and grid contents.
- Spawn coordinates are valid and within the map boundaries.
- Rooms and objects, when present, satisfy the game's schema.
- Empty cells use the exact representation expected by the Go loader.
- All required fields and data types are present.

Validation errors should identify the relevant field, tile, layer, or coordinate whenever possible.

### 5.3 Download

On successful validation, generate a JSON file in the browser and initiate a local download. The exported file must be usable directly as a game map file without manual editing or conversion.

## 6. Technical stack

### 6.1 Core technologies

- **Framework:** SvelteKit.
- **Language:** TypeScript.
- **UI:** Svelte components, HTML, and CSS.
- **Map rendering:** HTML Canvas 2D API.
- **Data serialization:** Native JavaScript JSON serialization, with explicit schema validation.
- **File import and export:** Browser File API, image decoding, Blob, and object URLs.

The initial implementation should be client-side and should not require a database or application server for editing and exporting maps.

### 6.2 Rendering architecture

Use ordinary HTML and CSS for the interface and Canvas 2D for the map workspace.

The canvas should be responsible for:

- Drawing visible map tiles.
- Drawing grid lines and workspace overlays.
- Drawing the selected-tile placement preview.
- Translating pointer coordinates into map coordinates.
- Handling zoom and panning transformations.
- Redrawing affected regions when practical.

The Svelte interface should be responsible for:

- Panel layout and navigation.
- Forms and configuration inputs.
- Tile selection and property editing.
- Layer management controls.
- Validation messages and export actions.

The canvas must not be the authoritative storage for map contents. It is a rendering surface backed by the application's map model.

### 6.3 Suggested module boundaries

Keep the application organized into separate modules with clear responsibilities:

- `map-model`: Types and in-memory representation of maps, tiles, layers, rooms, and objects.
- `tileset`: Image loading, source dimensions, slicing, and tile selection.
- `renderer`: Canvas drawing, grid rendering, previews, and viewport transformations.
- `editor-input`: Pointer handling, painting, erasing, zoom, and panning.
- `map-operations`: Cell updates, layer operations, and map resizing.
- `map-validation`: Structural and semantic validation.
- `map-serializer`: Conversion to the exact game JSON schema.
- `file-io`: Tileset import and JSON download.
- `ui`: Svelte components for panels, forms, tabs, and controls.

These are logical module boundaries, not mandatory directory names. Components should not duplicate map-editing logic or implement serialization independently.

### 6.4 Rendering performance

The implementation should:

- Draw only the portion of the map visible in the viewport.
- Avoid creating one DOM element per map cell.
- Avoid redrawing the complete map for every UI state change when unnecessary.
- Keep input handling responsive during continuous painting.
- Separate workspace rendering from inspector and panel updates.
- Use requestAnimationFrame where appropriate for scheduled rendering.

Canvas 2D is sufficient for the initial editor. A more advanced renderer should only be considered if measured performance demonstrates a need.

### 6.5 State management

Start with Svelte's built-in reactive state and TypeScript types.

Keep the map model separate from transient editor state. Transient state includes the active tile, active layer, selected tab, zoom, pan offset, and hover position.

Map mutations should go through dedicated operations so that painting, erasing, resizing, and layer changes remain consistent.

Undo snapshots the map and editor-side tile bindings before each committed operation. Drag strokes are grouped into a single step. Redo replays undone snapshots and is discarded when a new edit is made.

### 6.6 Dependencies

Prefer browser APIs and Svelte's built-in capabilities where practical. Add external dependencies only when they solve a concrete problem, such as schema validation or a substantial usability requirement.

A canvas rendering library is not required for the initial version.

## 7. Reliability and usability

- Handle invalid image files and malformed map data without crashing.
- Warn users before destructive operations such as deleting populated layers or shrinking a map.
- Keep layer visibility changes non-destructive.
- Ensure zoom and panning do not cause coordinate drift.
- Prevent invalid map references from reaching the exported file.
- Display useful errors rather than silently ignoring invalid operations.
- Keep the editor usable at different desktop viewport sizes.
- Avoid unnecessary server dependencies for local editing.

## 8. Initial scope and future improvements

### Initial release

The initial release should focus on the complete core workflow:

1. Configure a map.
2. Import and slice a tileset.
3. Select tiles and edit their properties.
4. Paint and erase tiles.
5. Manage layer visibility and selection.
6. Configure the spawn point.
7. Validate and export game-compatible JSON.

### Future improvements

Potential extensions include:

- Save and reopen editable projects.
- Keyboard shortcuts.
- Tile search and filtering.
- Fill and rectangle-paint tools.
- Copy and paste regions.
- Object placement and editing.
- Room editing.
- Map resizing with content preservation.
- Export previews and schema diagnostics.

These features should not delay delivery of the core editing and export workflow.

## 9. Acceptance criteria

The initial version is complete when:

- A user can create a map with configurable dimensions.
- A user can import an image and configure how it is sliced into tiles.
- A user can select a tile and see its preview and properties.
- A user can paint, drag-paint, and erase cells on the map.
- The workspace supports zooming and panning without changing logical map coordinates.
- Multiple layers can be selected and their visibility toggled independently.
- Map metadata and the spawn point can be configured.
- The editor rejects invalid map references and other schema violations.
- The exported JSON conforms to the existing game schema and loads successfully in the Go game.
- The editor remains responsive for maps of the sizes required by the game.
