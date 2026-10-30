package dialogue

type DialogueNode struct {
	Speaker string  `json:"speaker"`
	Text    string  `json:"text"`
	Next    *string `json:"next"`
}

type Dialogue struct {
	ID    string                  `json:"id"`
	Start string                  `json:"start"`
	Nodes map[string]DialogueNode `json:"nodes"`
}

type Catalog map[string]*Dialogue
