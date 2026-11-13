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
