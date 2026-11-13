package cutscene

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadDir(dir string) (Catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cutscene: read %s: %w", dir, err)
	}

	catalog := make(Catalog, len(entries))
	sources := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())

		cs, err := loadFile(path)
		if err != nil {
			return nil, err
		}

		if previous, exists := sources[cs.ID]; exists {
			return nil, fmt.Errorf("cutscene: duplicate id %q in %s and %s", cs.ID, previous, path)
		}
		sources[cs.ID] = path
		catalog[cs.ID] = cs
	}

	return catalog, nil
}

func loadFile(path string) (*Cutscene, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cutscene: read %s: %w", path, err)
	}

	var cs Cutscene
	if err := json.Unmarshal(data, &cs); err != nil {
		return nil, fmt.Errorf("cutscene: parse %s: %w", path, err)
	}

	if err := validate(path, &cs); err != nil {
		return nil, err
	}

	return &cs, nil
}

func validate(path string, cs *Cutscene) error {
	if cs.ID == "" {
		return fmt.Errorf("cutscene: %s: missing id", path)
	}
	if len(cs.Steps) == 0 {
		return fmt.Errorf("cutscene: %s: cutscene %q has no steps", path, cs.ID)
	}

	for i, step := range cs.Steps {
		if err := validateStep(path, cs.ID, i, step); err != nil {
			return err
		}
	}

	return nil
}

func validateStep(path, cutsceneID string, index int, step Step) error {
	switch step.Type {
	case StepDialogue:
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("cutscene: %s: cutscene %q: step %d: dialogue step requires non-empty id", path, cutsceneID, index)
		}
	case StepWait:
		if step.Duration <= 0 {
			return fmt.Errorf("cutscene: %s: cutscene %q: step %d: wait step requires duration > 0", path, cutsceneID, index)
		}
	case "":
		return fmt.Errorf("cutscene: %s: cutscene %q: step %d: missing type", path, cutsceneID, index)
	default:
		return fmt.Errorf("cutscene: %s: cutscene %q: step %d: unknown step type %q", path, cutsceneID, index, step.Type)
	}
	return nil
}
