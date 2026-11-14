package world

import rl "github.com/gen2brain/raylib-go/raylib"

type ActivationKind int

const (
	ActivationNone ActivationKind = iota
	ActivationDialogue
	ActivationCutscene
	ActivationSit
	ActivationStand
)

type ActivationOutcome struct {
	Kind       ActivationKind
	DialogueID string
	CutsceneID string
	SeatBounds rl.Rectangle
}

type ResolvedInteraction struct {
	Target      Interactable
	Prompt      string
	CanActivate bool
}

func ResolveInteraction(m *MapData, playerRect rl.Rectangle, playerCenter rl.Vector2, facing rl.Vector2) ResolvedInteraction {
	target := m.FindInteractable(playerRect, playerCenter, facing)
	if target == nil {
		return ResolvedInteraction{}
	}

	prompt := target.Prompt()
	canActivate := true
	if aware, ok := target.(PlayerAwareInteractable); ok {
		prompt = aware.PromptFor(playerRect)
		canActivate = aware.CanActivate(playerRect)
	}

	return ResolvedInteraction{
		Target:      target,
		Prompt:      prompt,
		CanActivate: canActivate,
	}
}

// SeatState exposes seat-specific state needed to derive sit/stand outcomes
// after Interact() has toggled occupancy.
type SeatState interface {
	IsOccupied() bool
	SeatBounds() rl.Rectangle
}

func ActivateInteraction(target Interactable, playerRect rl.Rectangle, playerSitting bool) ActivationOutcome {
	if target == nil {
		return ActivationOutcome{Kind: ActivationNone}
	}

	canActivate := true
	if aware, ok := target.(PlayerAwareInteractable); ok {
		canActivate = aware.CanActivate(playerRect)
	}
	if !canActivate {
		return ActivationOutcome{Kind: ActivationNone}
	}

	target.Interact()

	if talker, ok := target.(DialogueInteractable); ok {
		if id := talker.DialogueID(); id != "" {
			return ActivationOutcome{Kind: ActivationDialogue, DialogueID: id}
		}
	}

	if seat, ok := target.(SeatState); ok {
		if seat.IsOccupied() && !playerSitting {
			return ActivationOutcome{Kind: ActivationSit, SeatBounds: seat.SeatBounds()}
		}
		if !seat.IsOccupied() && playerSitting {
			return ActivationOutcome{Kind: ActivationStand}
		}
	}

	if cs, ok := target.(CutsceneInteractable); ok {
		if id := cs.CutsceneID(); id != "" {
			return ActivationOutcome{Kind: ActivationCutscene, CutsceneID: id}
		}
	}

	return ActivationOutcome{Kind: ActivationNone}
}
