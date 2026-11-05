package world

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// closestPointOnRect returns the point of the rectangle nearest to point.
// Points already inside the rectangle are returned unchanged.
func closestPointOnRect(rect rl.Rectangle, point rl.Vector2) rl.Vector2 {
	x := point.X
	if x < rect.X {
		x = rect.X
	} else if x > rect.X+rect.Width {
		x = rect.X + rect.Width
	}

	y := point.Y
	if y < rect.Y {
		y = rect.Y
	} else if y > rect.Y+rect.Height {
		y = rect.Y + rect.Height
	}

	return rl.Vector2{X: x, Y: y}
}

func pointInsideRect(rect rl.Rectangle, point rl.Vector2) bool {
	return point.X >= rect.X &&
		point.X <= rect.X+rect.Width &&
		point.Y >= rect.Y &&
		point.Y <= rect.Y+rect.Height
}

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
		targetPoint := closestPointOnRect(targetRect, playerCenter)
		toTarget := rl.Vector2{
			X: targetPoint.X - playerCenter.X,
			Y: targetPoint.Y - playerCenter.Y,
		}

		overlapping := pointInsideRect(targetRect, playerCenter)
		bodyOverlap := !object.Solid() &&
			rl.CheckCollisionRecs(playerRect, object.CollisionRect())

		if !overlapping && !bodyOverlap && facingSqrLength > 0 && toTarget.LengthSqr() > 0 {
			dot := rl.Vector2DotProduct(rl.Vector2Normalize(toTarget), facing)
			if dot < interactionFacingThreshold {
				continue
			}
		}

		// An overlapping player has zero distance by definition, so the
		// closest point also ranks wide objects fairly at their edges.
		distanceSquared := rl.Vector2DistanceSqr(playerCenter, targetPoint)
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
