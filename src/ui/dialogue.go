package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/dialogue"
)

var (
	speakerColor = rl.NewColor(233, 196, 106, 255)
	textColor    = rl.NewColor(224, 224, 214, 255)
	panelColor   = rl.NewColor(18, 20, 26, 235)
	borderColor  = rl.NewColor(140, 140, 140, 255)
)

func Measure(text string, fontSize float32) float32 {
	return rl.MeasureTextEx(rl.GetFontDefault(), text, fontSize, 1).X
}

func toRect(r dialogue.Rect) rl.Rectangle {
	return rl.Rectangle{
		X:      r.X,
		Y:      r.Y,
		Width:  r.Width,
		Height: r.Height,
	}
}

func DrawDialogue(snap dialogue.Snapshot, l dialogue.Layout) {
	font := rl.GetFontDefault()

	rl.DrawRectangleRec(toRect(l.Panel), panelColor)
	rl.DrawRectangleLinesEx(toRect(l.Panel), 2, borderColor)

	if snap.Speaker != "" && l.Speaker.Width > 0 {
		rl.DrawTextEx(font, snap.Speaker,
			rl.Vector2{X: l.Speaker.X, Y: l.Speaker.Y},
			l.FontSize, 1, speakerColor)
	}

	remaining := len([]rune(snap.Visible))
	y := l.Text.Y
	page := pageLines(snap, l)
	for _, line := range page {
		if remaining <= 0 {
			break
		}
		runes := []rune(line.Text)
		drawn := runes
		if remaining < len(runes) {
			drawn = runes[:remaining]
		}
		rl.DrawTextEx(font, string(drawn),
			rl.Vector2{X: l.Text.X, Y: y},
			l.FontSize, 1, textColor)
		remaining -= len(runes)
		y += l.LineHeight
	}
}

func pageLines(snap dialogue.Snapshot, l dialogue.Layout) []dialogue.Line {
	if snap.PageIndex < 0 || snap.PageIndex >= len(l.Pages) {
		return nil
	}
	return l.Pages[snap.PageIndex].Lines
}
