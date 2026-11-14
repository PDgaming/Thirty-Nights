package game

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	transitionFadeOutDuration = 0.35
	transitionFadeInDuration  = 0.35
)

type TransitionPhase int

const (
	TransitionNone TransitionPhase = iota
	TransitionFadeOut
	TransitionLoad
	TransitionFadeIn
)

func (g *Game) TransitionActive() bool {
	return g.transitionPhase != TransitionNone
}

func (g *Game) TransitionAlpha() float32 {
	return g.transitionAlpha
}

func (g *Game) RequestLevelTransition(target string) {
	if target == "" || g.TransitionActive() {
		return
	}
	g.transitionTarget = target
	g.transitionPhase = TransitionFadeOut
	g.transitionAlpha = 0
	g.transitionElapsed = 0
}

func (g *Game) updateTransition(dt float32) {
	switch g.transitionPhase {
	case TransitionNone:
		return
	case TransitionFadeOut:
		g.transitionElapsed += dt
		g.transitionAlpha = clamp01(g.transitionElapsed / transitionFadeOutDuration)
		if g.transitionAlpha >= 1 {
			g.transitionAlpha = 1
			g.transitionPhase = TransitionLoad
		}
	case TransitionLoad:
		g.performTransitionLoad()
	case TransitionFadeIn:
		g.transitionElapsed += dt
		g.transitionAlpha = 1 - clamp01(g.transitionElapsed/transitionFadeInDuration)
		if g.transitionAlpha <= 0 {
			g.completeTransition()
		}
	}
}

func (g *Game) performTransitionLoad() {
	target := g.transitionTarget
	g.transitionTarget = ""
	// Load while fully faded to black. Preserve current behavior: ignore load
	// errors here and proceed to fade-in to avoid trapping the player.
	// Startup/content validation should prevent this path for known content.
	_ = g.LoadLevel(target)
	g.transitionPhase = TransitionFadeIn
	g.transitionElapsed = 0
	g.transitionAlpha = 1
}

func (g *Game) completeTransition() {
	g.transitionAlpha = 0
	g.transitionPhase = TransitionNone
	g.transitionElapsed = 0
	if g.Map != nil && g.Map.StartCutscene != "" {
		g.startCutscene(g.Map.StartCutscene)
	}
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func (g *Game) DrawTransitionOverlay() {
	if g.transitionAlpha <= 0 {
		return
	}
	a := uint8(g.transitionAlpha * 255)
	if a == 0 {
		return
	}
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Color{R: 0, G: 0, B: 0, A: a})
}
