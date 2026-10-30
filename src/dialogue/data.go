package dialogue

const DefaultSpeed float32 = 40

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

type Page struct {
	Start int
	End   int
}

type Session struct {
	def      *Dialogue
	node     string
	runes    []rune
	pages    []Page
	page     int
	revealed float32
	speed    float32
	finished bool
}

type Snapshot struct {
	Node       string
	Speaker    string
	PageText   string
	Visible    string
	PageIndex  int
	PageCount  int
	RevealDone bool
	Finished   bool
}
