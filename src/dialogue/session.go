package dialogue

func NewSession(def *Dialogue) *Session {
	s := &Session{def: def, speed: DefaultSpeed}
	if def == nil {
		s.finished = true
		return s
	}
	s.enterNode(def.Start)
	return s
}

func (s *Session) SetSpeed(runesPerSecond float32) {
	if runesPerSecond > 0 {
		s.speed = runesPerSecond
	}
}

func (s *Session) Update(dt float32) {
	if s.finished || len(s.pages) == 0 || dt <= 0 {
		return
	}
	end := float32(s.pages[s.page].End)
	if s.revealed >= end {
		return
	}
	s.revealed = min(s.revealed+s.speed*dt, end)
}

func (s *Session) Advance() {
	if s.finished || s.def == nil || len(s.pages) == 0 {
		return
	}

	page := s.pages[s.page]
	if s.revealed < float32(page.End) {
		s.revealed = float32(page.End)
		return
	}

	if s.page < len(s.pages)-1 {
		s.page++
		s.revealed = float32(s.pages[s.page].Start)
		return
	}

	node := s.def.Nodes[s.node]
	if node.Next == nil || *node.Next == "" {
		s.finished = true
		return
	}

	s.enterNode(*node.Next)
}

func (s *Session) SetPages(pages []Page) {
	if len(s.runes) == 0 || len(pages) == 0 {
		return
	}

	clean := make([]Page, 0, len(pages))
	for _, p := range pages {
		start := max(p.Start, 0)
		end := min(p.End, len(s.runes))
		if end <= start {
			continue
		}
		clean = append(clean, Page{Start: start, End: end})
	}
	if len(clean) == 0 {
		return
	}

	clean[len(clean)-1].End = len(s.runes)

	s.pages = clean
	if s.page >= len(clean) {
		s.page = len(clean) - 1
	}
	if s.revealed < float32(clean[s.page].Start) {
		s.revealed = float32(clean[s.page].Start)
	}
}

func (s *Session) Snapshot() Snapshot {
	snap := Snapshot{
		Finished:  s.finished,
		PageIndex: s.page,
		PageCount: len(s.pages),
	}
	if s.finished || s.def == nil || len(s.pages) == 0 {
		return snap
	}

	page := s.pages[s.page]
	node := s.def.Nodes[s.node]

	snap.Node = s.node
	snap.Speaker = node.Speaker
	snap.PageText = string(s.runes[page.Start:page.End])
	end := min(int(s.revealed), page.End)
	if end > page.Start {
		snap.Visible = string(s.runes[page.Start:end])
	}
	snap.RevealDone = int(s.revealed) >= page.End
	return snap
}

func (s *Session) enterNode(key string) {
	node, ok := s.def.Nodes[key]
	if !ok {
		s.finished = true
		return
	}
	s.node = key
	s.runes = []rune(node.Text)
	s.pages = []Page{{Start: 0, End: len(s.runes)}}
	s.page = 0
	s.revealed = 0
	s.finished = false
}
