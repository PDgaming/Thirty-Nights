package world

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *MapData) FindInteractable(playerRect rl.Rectangle, playerCenter rl.Vector2, facing rl.Vector2) Interactable {
	if facing.Length() != 0 {
		facing = rl.Vector2Normalize(facing)
	}
	facingSqrLength := facing.LengthSqr()

	var best Interactable
	bestDistanceSquared := float32(math.Inf(1))

	for _, objectData := range m.Objects {
		object, ok := m.ObjectsByID[objectData.ID]
		if !ok {
			continue
		}
		target, ok := object.(Interactable)
		if !ok {
			continue
		}

		if !target.CanInteractRect(playerRect) {
			continue
		}

		targetRect := target.InteractionRect()
		targetCenter := rl.Vector2{
			X: targetRect.X + targetRect.Width/2,
			Y: targetRect.Y + targetRect.Height/2,
		}
		toTarget := rl.Vector2{
			X: targetCenter.X - playerCenter.X,
			Y: targetCenter.Y - playerCenter.Y,
		}
		if toTarget.Length() == 0 {
			return target
		}

		bodyOverlap := !object.Solid() &&
			rl.CheckCollisionRecs(playerRect, object.CollisionRect())

		if !bodyOverlap && facingSqrLength > 0 {
			dot := rl.Vector2DotProduct(rl.Vector2Normalize(toTarget), facing)
			if dot < interactionFacingThreshold {
				continue
			}
		}

		distanceSquared := rl.Vector2DistanceSqr(playerCenter, targetCenter)
		if distanceSquared < bestDistanceSquared {
			best = target
			bestDistanceSquared = distanceSquared
		}
	}
	if best == nil {
		return nil
	}
	return best
}
