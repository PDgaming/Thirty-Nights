package cutscene

type StepType string

const (
	StepDialogue StepType = "dialogue"
	StepWait     StepType = "wait"
)

type Step struct {
	Type     StepType `json:"type"`
	Duration float32  `json:"duration,omitempty"`
	ID       string   `json:"id,omitempty"`
}

type Cutscene struct {
	ID    string `json:"id"`
	Steps []Step `json:"steps"`
}

type Catalog map[string]*Cutscene

func (c Catalog) Get(id string) (*Cutscene, bool) {
	cs, ok := c[id]
	return cs, ok
}

func DialogueIDs(cs *Cutscene) []string {
	if cs == nil {
		return nil
	}
	ids := make([]string, 0, len(cs.Steps))
	for _, step := range cs.Steps {
		if step.Type == StepDialogue && step.ID != "" {
			ids = append(ids, step.ID)
		}
	}
	return ids
}
