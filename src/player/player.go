package player

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/world"
)

func New(x, y float32) *Player {
	p := &Player{
		Position: rl.Vector2{X: x, Y: y},
		Speed:    100,
		Size:     rl.Vector2{X: 12, Y: 14},
		Facing:   rl.Vector2{X: -1, Y: 0},
		Animator: *NewAnimator(),
		Sitting:  false,
	}
	return p
}

func (p *Player) Update(dt float32, m *world.MapData) {
	movement := rl.Vector2{
		X: float32(0),
		Y: float32(0),
	}

	if rl.IsKeyDown(rl.KeyW) {
		movement.Y -= 1
	}
	if rl.IsKeyDown(rl.KeyA) {
		movement.X -= 1
	}
	if rl.IsKeyDown(rl.KeyS) {
		movement.Y += 1
	}
	if rl.IsKeyDown(rl.KeyD) {
		movement.X += 1
	}

	if rl.Vector2Length(movement) > 0 {
		p.Animator.ChangeState(Walking)
	} else {
		p.Animator.ChangeState(Idle)
	}

	p.updateFacing(movement)

	if p.Sitting {
		p.Animator.ChangeState(Idle)
		return
	}

	movement = rl.Vector2Normalize(movement)

	hDisplacement := movement.X * p.Speed * dt
	vDisplacement := movement.Y * p.Speed * dt

	p.CollisionChecks(hDisplacement, vDisplacement, m)
}

func (p *Player) Draw(debug bool) {
	switch p.Animator.State {
	case Idle:
		rl.DrawRectangle(int32(p.Position.X), int32(p.Position.Y), int32(p.Size.X), int32(p.Size.Y), rl.Green)
	case Walking:
		switch p.Animator.Frame { // Hardcoded number of frames for walking animation
		case 0:
			rl.DrawRectangle(int32(p.Position.X), int32(p.Position.Y), int32(p.Size.X), int32(p.Size.Y), rl.Red)
		case 1:
			rl.DrawRectangle(int32(p.Position.X), int32(p.Position.Y), int32(p.Size.X), int32(p.Size.Y), rl.Yellow)
		}
	}

	if debug {
		rl.DrawRectangleLines(int32(p.Position.X), int32(p.Position.Y), int32(p.Size.X), int32(p.Size.Y), rl.Black)

		center := p.Center()
		rl.DrawLine(int32(center.X), int32(center.Y), int32(center.X+p.Facing.X*10), int32(center.Y+p.Facing.Y*10), rl.Red)
	}
}

func (p *Player) CollisionChecks(hDisplacement, vDisplacement float32, m *world.MapData) {
	p.Position.X += hDisplacement
	collisionRect := rl.Rectangle{
		X:      p.Position.X,
		Y:      p.Position.Y,
		Width:  p.Size.X,
		Height: p.Size.Y,
	}

	if m.CanCollidRect(collisionRect) || m.CanCollideObjects(collisionRect) {
		p.Position.X -= hDisplacement
	}

	p.Position.Y += vDisplacement
	collisionRect.X = p.Position.X
	collisionRect.Y = p.Position.Y

	if m.CanCollidRect(collisionRect) || m.CanCollideObjects(collisionRect) {
		p.Position.Y -= vDisplacement
	}
}

// Helper functions
func (p *Player) updateFacing(movement rl.Vector2) {
	if movement.X == 0 && movement.Y == 0 {
		return
	}
	if math.Abs(float64(movement.X)) == math.Abs(float64(movement.Y)) {
		return
	}
	if math.Abs(float64(movement.X)) > math.Abs(float64(movement.Y)) {
		if movement.X > 0 {
			p.Facing = rl.Vector2{X: 1, Y: 0}
		} else {
			p.Facing = rl.Vector2{X: -1, Y: 0}
		}
	} else {
		if movement.Y > 0 {
			p.Facing = rl.Vector2{X: 0, Y: 1}
		} else {
			p.Facing = rl.Vector2{X: 0, Y: -1}
		}
	}
}
func (p *Player) Rect() rl.Rectangle {
	rect := rl.Rectangle{
		X:      p.Position.X,
		Y:      p.Position.Y,
		Width:  p.Size.X,
		Height: p.Size.Y,
	}
	return rect
}
func (p *Player) Center() rl.Vector2 {
	return rl.Vector2{
		X: p.Position.X + p.Size.X/2,
		Y: p.Position.Y + p.Size.Y/2,
	}
}

func (p *Player) Sit(seatRect rl.Rectangle) {
	p.savedPosition = p.Position
	p.Position = rl.Vector2{
		X: seatRect.X + (seatRect.Width-p.Size.X)/2,
		Y: seatRect.Y + (seatRect.Height-p.Size.Y)/2,
	}
	p.Sitting = true
}
func (p *Player) Stand() {
	p.Position = p.savedPosition
	p.Sitting = false
}
