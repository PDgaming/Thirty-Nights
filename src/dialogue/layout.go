package dialogue

import "math"

func (r Rect) Right() float32  { return r.X + r.Width }
func (r Rect) Bottom() float32 { return r.Y + r.Height }

func (l Layout) Ranges() []Page {
	ranges := make([]Page, 0, len(l.Pages))
	for _, page := range l.Pages {
		if len(page.Lines) == 0 {
			ranges = append(ranges, Page{})
			continue
		}
		ranges = append(ranges, Page{
			Start: page.Lines[0].Start,
			End:   page.Lines[len(page.Lines)-1].End,
		})
	}
	return ranges
}

func Compute(screenWidth, screenHeight float32, measure MeasureFunc, speaker, text string) Layout {
	if screenWidth <= 0 || screenHeight <= 0 || measure == nil {
		return Layout{}
	}

	scale := min(max(screenHeight/450, 0.7), 2.0)
	fontSize := round(18 * scale)
	lineHeight := round(fontSize * 1.45)
	padding := round(10 * scale)
	gap := round(6 * scale)
	margin := round(12 * scale)

	panelWidth := min(screenWidth*0.9, 900)
	panelWidth = max(panelWidth, 300)
	panelWidth = min(panelWidth, screenWidth-2*margin)
	if panelWidth < 1 {
		return Layout{}
	}

	panelHeight := max(screenHeight*0.38, 110)
	panelHeight = min(panelHeight, screenHeight-2*margin)
	if panelHeight < 1 {
		return Layout{}
	}

	panel := Rect{
		X:      round((screenWidth - panelWidth) / 2),
		Y:      round(screenHeight - margin - panelHeight),
		Width:  round(panelWidth),
		Height: round(panelHeight),
	}

	inner := Rect{
		X:      panel.X + padding,
		Y:      panel.Y + padding,
		Width:  panel.Width - 2*padding,
		Height: panel.Height - 2*padding,
	}

	speakerRect := Rect{}
	textTop := inner.Y
	if speaker != "" {
		speakerRect = Rect{X: inner.X, Y: inner.Y, Width: inner.Width, Height: lineHeight}
		textTop = speakerRect.Y + lineHeight + gap
	}

	indicatorSize := round(fontSize * 0.9)
	indicator := Rect{
		X:      inner.X + inner.Width - indicatorSize,
		Y:      inner.Y + inner.Height - indicatorSize,
		Width:  indicatorSize,
		Height: indicatorSize,
	}

	textRect := Rect{
		X:      inner.X,
		Y:      textTop,
		Width:  inner.Width,
		Height: indicator.Y - gap - textTop,
	}

	lines := wrapLines([]rune(text), textRect.Width, measure, fontSize)
	maxLines := int(textRect.Height / lineHeight)
	if maxLines < 1 {
		maxLines = 1
	}

	return Layout{
		Metrics: Metrics{
			Panel:      panel,
			Speaker:    speakerRect,
			Text:       textRect,
			Indicator:  indicator,
			FontSize:   fontSize,
			LineHeight: lineHeight,
			Padding:    padding,
		},
		Pages: paginate(lines, maxLines),
	}
}

func wrapLines(runes []rune, maxWidth float32, measure MeasureFunc, fontSize float32) []Line {
	if len(runes) == 0 {
		return []Line{{Start: 0, End: 0}}
	}

	var lines []Line
	start := 0
	for start < len(runes) {
		end := start
		lastSpace := -1

		for end < len(runes) && runes[end] != '\n' {
			if end > start && measure(string(runes[start:end+1]), fontSize) > maxWidth {
				break
			}
			if runes[end] == ' ' {
				lastSpace = end
			}
			end++
		}

		lineEnd := end
		if end < len(runes) && runes[end] == '\n' {
			lineEnd = end + 1
		} else if end < len(runes) && lastSpace >= start {
			lineEnd = lastSpace + 1
		}
		if lineEnd <= start {
			lineEnd = start + 1
		}

		lines = append(lines, Line{
			Text:  string(runes[start:lineEnd]),
			Start: start,
			End:   lineEnd,
		})
		start = lineEnd
	}

	return lines
}

func paginate(lines []Line, perPage int) []TextPage {
	if perPage < 1 {
		perPage = 1
	}

	pages := make([]TextPage, 0, (len(lines)+perPage-1)/perPage)
	for start := 0; start < len(lines); start += perPage {
		end := min(start+perPage, len(lines))
		page := TextPage{Lines: make([]Line, end-start)}
		copy(page.Lines, lines[start:end])
		pages = append(pages, page)
	}
	return pages
}

func round(v float32) float32 {
	return float32(math.Round(float64(v)))
}
