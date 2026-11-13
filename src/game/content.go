package game

import (
	"fmt"

	"thirty.nights/src/cutscene"
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

func CheckCutscenes(m *world.MapData, catalog cutscene.Catalog) error {
	if m.StartCutscene != "" {
		if _, ok := catalog[m.StartCutscene]; !ok {
			return fmt.Errorf("map %q requests unknown start_cutscene %q", m.ID, m.StartCutscene)
		}
	}

	for _, object := range m.Objects {
		obj, ok := m.ObjectsByID[object.ID]
		if !ok {
			continue
		}
		cs, ok := obj.(world.CutsceneInteractable)
		if !ok {
			continue
		}
		id := cs.CutsceneID()
		if id == "" {
			continue
		}
		if _, ok := catalog[id]; !ok {
			return fmt.Errorf("object %q requests unknown cutscene %q", object.ID, id)
		}
	}
	return nil
}
