package cutscene

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCutsceneFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write cutscene file: %v", err)
	}
	return path
}

func TestLoadDirValidCutscene(t *testing.T) {
	dir := t.TempDir()
	writeCutsceneFile(t, dir, "intro.json", `{
  "id": "intro",
  "steps": [
    {"type": "wait", "duration": 0.5},
    {"type": "dialogue", "id": "receptionist"}
  ]
}`)

	catalog, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir() error = %v", err)
	}
	cs, ok := catalog.Get("intro")
	if !ok {
		t.Fatal("catalog missing cutscene \"intro\"")
	}
	if len(cs.Steps) != 2 {
		t.Fatalf("len(steps) = %d, want 2", len(cs.Steps))
	}
	if cs.Steps[0].Type != StepWait || cs.Steps[0].Duration != 0.5 {
		t.Fatalf("step 0 = %+v, want wait 0.5", cs.Steps[0])
	}
	if cs.Steps[1].Type != StepDialogue || cs.Steps[1].ID != "receptionist" {
		t.Fatalf("step 1 = %+v, want dialogue receptionist", cs.Steps[1])
	}
}

func TestLoadDirRejectsInvalidContent(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		content  string
		contains string
	}{
		{name: "malformed json", fileName: "bad.json", content: `{"id":`, contains: "parse"},
		{name: "missing id", fileName: "noid.json", content: `{"steps":[{"type":"wait","duration":1}]}`, contains: "missing id"},
		{name: "no steps", fileName: "nosteps.json", content: `{"id":"x","steps":[]}`, contains: "has no steps"},
		{name: "missing step type", fileName: "notype.json", content: `{"id":"x","steps":[{"duration":1}]}`, contains: "missing type"},
		{name: "unknown step type", fileName: "unknown.json", content: `{"id":"x","steps":[{"type":"camera"}]}`, contains: "unknown step type"},
		{name: "wait without duration", fileName: "waitbad.json", content: `{"id":"x","steps":[{"type":"wait","duration":0}]}`, contains: "duration > 0"},
		{name: "dialogue without id", fileName: "dlg.json", content: `{"id":"x","steps":[{"type":"dialogue"}]}`, contains: "requires non-empty id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeCutsceneFile(t, dir, tt.fileName, tt.content)
			_, err := LoadDir(dir)
			if err == nil {
				t.Fatal("LoadDir() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("LoadDir() error = %q, want to contain %q", err, tt.contains)
			}
		})
	}
}

func TestLoadDirRejectsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	content := `{"id":"dup","steps":[{"type":"wait","duration":1}]}`
	writeCutsceneFile(t, dir, "a.json", content)
	writeCutsceneFile(t, dir, "b.json", content)

	_, err := LoadDir(dir)
	if err == nil {
		t.Fatal("LoadDir() error = nil, want duplicate error")
	}
	if !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("LoadDir() error = %q, want duplicate id", err)
	}
}
