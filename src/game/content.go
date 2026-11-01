package game

import (
	"fmt"

	"thirty.nights/src/dialogue"
	"thirty.nights/src/world"
)

func CheckDialogues(m *world.MapData, catalog dialogue.Catalog) error {
	for _, object := range m.Objects {
		obj, ok := m.ObjectsByID[object.ID]
		if !ok {
			continue
		}

		talker, ok := obj.(world.DialogueInteractable)
		if !ok {
			continue
		}

		id := talker.DialogueID()
		if id == "" {
			return fmt.Errorf("object %q requests an empty dialogue ID", object.ID)
		}
		if _, ok := catalog[id]; !ok {
			return fmt.Errorf("object %q requests unknown dialogue %q", object.ID, id)
		}
	}
	return nil
}
