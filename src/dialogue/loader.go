package dialogue

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
		return nil, fmt.Errorf("dialogue: read %s: %w", dir, err)
	}

	catalog := make(Catalog, len(entries))
	sources := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())

		d, err := loadFile(path)
		if err != nil {
			return nil, err
		}

		if previous, exists := sources[d.ID]; exists {
			return nil, fmt.Errorf("dialogue: duplicate id %q in %s and %s", d.ID, previous, path)
		}
		sources[d.ID] = path
		catalog[d.ID] = d
	}
	return catalog, nil
}

func loadFile(path string) (*Dialogue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dialogue: read %s: %w", path, err)
	}

	var d Dialogue
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("dialogue: parse %s: %w", path, err)
	}

	if err := validate(path, &d); err != nil {
		return nil, err
	}

	return &d, nil
}

func validate(path string, d *Dialogue) error {
	if d.ID == "" {
		return fmt.Errorf("dialogue: %s: missing id", path)
	}
	if d.Start == "" {
		return fmt.Errorf("dialogue: %s: dialogue %q has no start node", path, d.ID)
	}
	if len(d.Nodes) == 0 {
		return fmt.Errorf("dialogue: %s: dialogue %q has no nodes", path, d.ID)
	}
	if _, ok := d.Nodes[d.Start]; !ok {
		return fmt.Errorf("dialogue: %s: dialogue %q: start %q is not a node", path, d.ID, d.Start)
	}

	for key, node := range d.Nodes {
		if key == "" {
			return fmt.Errorf("dialogue: %s: dialogue %q has an empty node key", path, d.ID)
		}
		if strings.TrimSpace(node.Text) == "" {
			return fmt.Errorf("dialogue: %s: dialogue %q: node %q has no text", path, d.ID, key)
		}
		if node.Next == nil {
			continue
		}
		if *node.Next == "" {
			return fmt.Errorf("dialogue: %s: dialogue %q: node %q has an empty next", path, d.ID, key)
		}
		if _, ok := d.Nodes[*node.Next]; !ok {
			return fmt.Errorf("dialogue: %s: dialogue %q: node %q: next %q is not a node", path, d.ID, key, *node.Next)
		}
	}

	return nil
}
