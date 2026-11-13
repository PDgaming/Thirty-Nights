package game

import (
	"fmt"

	"thirty.nights/src/cutscene"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/world"
)

func CheckMapDependencies(catalog world.Catalog) error {
	for id, m := range catalog {
		for i, dep := range m.Maps {
			if _, ok := catalog[dep]; !ok {
				return fmt.Errorf("map %q: dependency maps[%d] references unknown map %q", id, i, dep)
			}
		}
	}
	return nil
}

func CheckExitTransitions(catalog world.Catalog) error {
	for mapID, m := range catalog {
		for _, object := range m.Objects {
			if object.Type != "exit" {
				continue
			}
			raw, ok := object.Properties["nextLevel"]
			if !ok || raw == nil {
				return fmt.Errorf("map %q: exit object %q missing nextLevel", mapID, object.ID)
			}
			nextLevel, ok := raw.(string)
			if !ok || nextLevel == "" {
				return fmt.Errorf("map %q: exit object %q has invalid nextLevel", mapID, object.ID)
			}
			if _, ok := catalog[nextLevel]; !ok {
				return fmt.Errorf("map %q: exit object %q references unknown map %q", mapID, object.ID, nextLevel)
			}
		}
	}
	return nil
}

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

func CheckCutscenes(m *world.MapData, cutscenes cutscene.Catalog, dialogues dialogue.Catalog) error {
	validateDialogueRefs := func(source string, cs *cutscene.Cutscene) error {
		for i, id := range cutscene.DialogueIDs(cs) {
			if _, ok := dialogues[id]; !ok {
				return fmt.Errorf("%s: cutscene %q step %d references unknown dialogue %q", source, cs.ID, i, id)
			}
		}
		return nil
	}

	if m.StartCutscene != "" {
		cs, ok := cutscenes[m.StartCutscene]
		if !ok {
			return fmt.Errorf("map %q requests unknown start_cutscene %q", m.ID, m.StartCutscene)
		}
		if err := validateDialogueRefs(fmt.Sprintf("map %q start_cutscene", m.ID), cs); err != nil {
			return err
		}
	}

	for _, object := range m.Objects {
		obj, ok := m.ObjectsByID[object.ID]
		if !ok {
			continue
		}
		trigger, ok := obj.(world.CutsceneInteractable)
		if !ok {
			continue
		}
		id := trigger.CutsceneID()
		if id == "" {
			continue
		}
		cs, ok := cutscenes[id]
		if !ok {
			return fmt.Errorf("object %q requests unknown cutscene %q", object.ID, id)
		}
		if err := validateDialogueRefs(fmt.Sprintf("object %q cutsceneId", object.ID), cs); err != nil {
			return err
		}
	}
	return nil
}
