package player

import (
	rl "github.com/gen2brain/raylib-go/raylib"
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

func (p *Player) Update(dt float32) {
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

	p.Position.X += movement.X * p.Speed * dt
	p.Position.Y += movement.Y * p.Speed * dt
}

func (p *Player) Draw() {
	rl.DrawRectangle(int32(p.Position.X), int32(p.Position.Y), int32(p.Size.X), int32(p.Size.Y), rl.Red)
}
