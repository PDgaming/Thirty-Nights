package dialogue

type DialogueNote struct {
	Speaker string  `json:"speaker"`
	Text    string  `json:"text"`
	Next    *string `json:"next"`
}

type Dialogue struct {
	ID    string                  `json:"id"`
	Start string                  `json:"start"`
	Nodes map[string]DialogueNote `json:"nodes"`
}

type Catalog map[string]*Dialogue
