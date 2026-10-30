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

type Line struct {
	Text  string
	Start int
	End   int
}

type TextPage struct {
	Lines []Line
}

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

type Rect struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

type Metrics struct {
	Panel      Rect
	Speaker    Rect
	Text       Rect
	Indicator  Rect
	FontSize   float32
	LineHeight float32
	Padding    float32
}

type Layout struct {
	Metrics
	Pages []TextPage
}

type MeasureFunc func(text string, fontSize float32) float32
