# Cutscene system

## Asset format

- Location: `asset/cutscenes/*.json`
- Loader: Go `src/cutscene.LoadDir` (parallel to `src/dialogue.LoadDir`)
- Each file defines one cutscene. File name does not determine ID.

### Cutscene JSON

```json
{
  "id": "office_intro",
  "steps": [
    { "type": "wait", "duration": 0.5 },
    { "type": "dialogue", "id": "receptionist" }
  ]
}
```

Fields:
- `id` (string, required): stable unique cutscene ID referenced by maps and objects.
- `steps` (array, required, non-empty): ordered steps executed sequentially.

### Step types (initial)

- `wait`
  - `duration` (number, required): seconds to wait. Must be > 0.
- `dialogue`
  - `id` (string, required): dialogue ID from the dialogue catalog. The runner drives the existing dialogue session and advances only after the dialogue finishes.

Unknown step types are invalid and rejected by the loader. Step-specific required fields are validated at load time.

## Trigger contract (shared map JSON)

### Level-start cutscene

- Map field: `start_cutscene` (string, optional)
- Meaning: cutscene ID to play when the level is entered/loaded.
- Absent, `null`, or empty string means no level-start cutscene.
- Backward compatible: existing maps without this field remain valid.

### Object post-interaction cutscene

- Object property key: `cutsceneId` (string, optional)
- Meaning: cutscene ID to play **after** a successful object interaction completes. The object's normal `Interact()` side effects (door toggle, seat state changes, etc.) run first; the cutscene starts after that interaction for the same input.
- Absent, `null`, or empty string means no post-interaction cutscene.
- Backward compatible: existing objects/properties remain valid.

No aliases are defined in the initial contract; `start_cutscene` and `cutsceneId` are canonical.

## Runtime ordering and behavior

- **Level start:** after a level/map is loaded and game state is initialized, if `start_cutscene` is set and valid, the cutscene runner starts and blocks normal play until completion.
- **Post-interaction:** when the player successfully activates an interactable, the engine performs the interaction, then checks for `cutsceneId`. If present, it starts that cutscene. If the interaction already started a dialogue through existing dialogue-interactable behavior, the dialogue takes precedence for that interaction and no additional cutscene is auto-started from the same trigger (cutscenes may contain dialogue steps instead).
- **Blocking:** while a cutscene is active, normal gameplay input/movement is blocked and interactables cannot be re-triggered. Cutscene-driven dialogue uses the existing dialogue presentation/advance input while the cutscene owns control.
- **Validation:** unknown cutscene IDs referenced by maps or objects fail content validation/load with errors identifying the source (map or object ID/property). Dialogue IDs referenced by cutscene `dialogue` steps are also validated against the loaded dialogue catalog during level content checks.

## TileMapEditor authoring

- **Level start:** set **Map Settings → Start cutscene** to a cutscene ID (or clear it to remove the trigger). The field round-trips as map-level `start_cutscene`.
- **Object trigger:** set an object's `cutsceneId` property (available on default door/npc/trigger/seat attributes, or add a string attribute named `cutsceneId`) to a cutscene ID. Empty/cleared values mean no trigger. Properties round-trip in the object's `properties` map.

## Compatibility notes

- Existing dialogue JSON and runtime are unchanged; dialogue steps reference dialogue IDs only.
- Adding `start_cutscene`/`cutsceneId` is additive and optional.
- Future step types may be added without breaking existing cutscenes as long as new types are explicitly supported by loader/runner.
- Current limitations: only `wait` and `dialogue` steps, no skip/branching/parallel tracks, and no in-editor cutscene preview.
