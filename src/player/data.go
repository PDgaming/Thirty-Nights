package player

import rl "github.com/gen2brain/raylib-go/raylib"

type AnimationState int

const (
	Idle AnimationState = iota
	Walking
)

type Player struct {
	Position rl.Vector2
	Speed    float32
	Size     rl.Vector2
	Animator Animator
}

type Animator struct {
	State         AnimationState
	Frame         int
	Elapsed       float32
	FrameDuration float32
}
