package player

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/world"
)

func New(x, y float32) *Player {
	p := &Player{
		Position: rl.Vector2{X: x, Y: y},
		Speed:    100,
		Size:     rl.Vector2{X: 12, Y: 14},
		Animator: *NewAnimator(),
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

	movement = rl.Vector2Normalize(movement)

	hDisplacement := movement.X * p.Speed * dt
	vDisplacement := movement.Y * p.Speed * dt

	p.Position.X += hDisplacement
	collisionRect := rl.Rectangle{
		X:      p.Position.X,
		Y:      p.Position.Y,
		Width:  p.Size.X,
		Height: p.Size.Y,
	}

	if m.CanCollidRect(collisionRect) {
		p.Position.X -= hDisplacement
	}

	p.Position.Y += vDisplacement
	collisionRect.X = p.Position.X
	collisionRect.Y = p.Position.Y

	if m.CanCollidRect(collisionRect) {
		p.Position.Y -= vDisplacement
	}
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
	}
}
