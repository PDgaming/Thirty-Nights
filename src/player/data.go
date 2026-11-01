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
	Facing   rl.Vector2
	Animator Animator

	Sitting       bool
	savedPosition rl.Vector2

	// Half-tile lattice movement state. logical is the body center and always
	// lies on the lattice (multiples of tileSize/2) while idle, so every
	// resting point is either a tile center or a tile edge. While stepping,
	// logical tweens from one lattice point to the next.
	tileSize float32
	logical  rl.Vector2
	stepping bool
	from     rl.Vector2
	to       rl.Vector2
	progress float32
	// Direction of the most recently pressed key, used to break ties when two
	// direction keys are held at once.
	lastHeld rl.Vector2
}

type Animation struct {
	FrameCount    int
	FrameDuration float32
	Loop          bool
}

type Animations map[AnimationState]Animation

type Animator struct {
	State      AnimationState
	Frame      int
	Elapsed    float32
	Animations Animations
}
