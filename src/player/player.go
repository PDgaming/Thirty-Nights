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
	p.logical = rl.Vector2{
		X: p.Position.X + p.Size.X/2,
		Y: p.Position.Y + p.Size.Y/2,
	}
	p.applyLogical(p.logical)
	return p
}

// readInput returns the cardinal direction keys currently held and records
// which key was pressed most recently, so simultaneous holds resolve
// predictably (the newest key wins instead of a fixed W > A > S > D order).
func readInput(p *Player) rl.Vector2 {
	input := rl.Vector2{X: 0, Y: 0}
	if rl.IsKeyDown(rl.KeyW) {
		input.Y -= 1
	}
	if rl.IsKeyDown(rl.KeyA) {
		input.X -= 1
	}
	if rl.IsKeyDown(rl.KeyS) {
		input.Y += 1
	}
	if rl.IsKeyDown(rl.KeyD) {
		input.X += 1
	}
	if rl.IsKeyPressed(rl.KeyW) {
		p.lastHeld = rl.Vector2{X: 0, Y: -1}
	}
	if rl.IsKeyPressed(rl.KeyA) {
		p.lastHeld = rl.Vector2{X: -1, Y: 0}
	}
	if rl.IsKeyPressed(rl.KeyS) {
		p.lastHeld = rl.Vector2{X: 0, Y: 1}
	}
	if rl.IsKeyPressed(rl.KeyD) {
		p.lastHeld = rl.Vector2{X: 1, Y: 0}
	}
	return input
}

// Update steps the player along a half-tile lattice: tile centers and tile
// edges are both valid resting points, so every step is one tile size per two
// steps (center -> edge -> center). A step commits to its target; turning
// happens on arrival, not mid-step. Input is re-read at every arrival, so a
// held key chains steps with no dead frame in between.
func (p *Player) Update(dt float32, m *world.MapData) {
	input := readInput(p)
	p.updateFacing(input)

	if p.Sitting {
		p.Animator.ChangeState(Idle)
		return
	}

	p.ensureAligned(m)

	if p.stepping {
		p.progress += p.Speed * dt / p.halfTile()
		if p.progress < 1 {
			p.applyLogical(rl.Vector2Lerp(p.from, p.to, p.progress))
			p.Animator.ChangeState(Walking)
			return
		}
		// Land exactly on the lattice point, then fall through to start the
		// next step from the leftover input.
		p.applyLogical(p.to)
		p.stepping = false
		p.progress = 0
	}

	if input.X == 0 && input.Y == 0 {
		p.Animator.ChangeState(Idle)
		return
	}
	p.Animator.ChangeState(Walking)

	for _, dir := range p.candidateSteps(input) {
		if p.canStep(m, dir) {
			p.startStep(dir)
			return
		}
	}
	// Every candidate step is blocked: stay put, keep the last frame's position.
}

// candidateSteps picks where a step may go, in priority order. With two
// direction keys held, the first candidate is the true diagonal step, followed
// by the two cardinals (most recently pressed key first) as slide fallbacks if
// the diagonal is walled off. Only straight input yields a single cardinal.
func (p *Player) candidateSteps(input rl.Vector2) []rl.Vector2 {
	if input.X != 0 && input.Y != 0 {
		last := p.lastHeld
		var firstCard, secondCard rl.Vector2
		switch {
		case last.X != 0:
			firstCard = rl.Vector2{X: input.X, Y: 0}
			secondCard = rl.Vector2{X: 0, Y: input.Y}
		case last.Y != 0:
			firstCard = rl.Vector2{X: 0, Y: input.Y}
			secondCard = rl.Vector2{X: input.X, Y: 0}
		default:
			// No press edge recorded yet (keys held since before this frame);
			// prefer horizontal so behaviour stays deterministic.
			firstCard = rl.Vector2{X: input.X, Y: 0}
			secondCard = rl.Vector2{X: 0, Y: input.Y}
		}
		return []rl.Vector2{
			{X: input.X, Y: input.Y},
			firstCard,
			secondCard,
		}
	}
	return []rl.Vector2{{X: input.X, Y: input.Y}}
}

func (p *Player) startStep(dir rl.Vector2) {
	h := p.halfTile()
	p.from = p.logical
	p.to = rl.Vector2{
		X: p.logical.X + dir.X*h,
		Y: p.logical.Y + dir.Y*h,
	}
	p.progress = 0
	p.stepping = true
}

// ensureAligned snaps the logical center to the half-tile lattice while idle.
// Every lattice point is either a tile center or a tile edge, so external
// position changes (spawn, standing up from a seat) land somewhere sensible.
func (p *Player) ensureAligned(m *world.MapData) {
	p.tileSize = float32(m.TileSize)
	if p.stepping {
		return
	}
	h := p.halfTile()
	p.applyLogical(rl.Vector2{
		X: float32(math.Round(float64(p.logical.X/h))) * h,
		Y: float32(math.Round(float64(p.logical.Y/h))) * h,
	})
}

// canStep reports whether a step in dir is legal. It checks the real body
// rectangle (12x14) centered on the candidate logical point, which restores
// the original collision semantics, now quantized to 8px moves.
//
// Cardinal steps only need the destination body clear: the swept region along
// the way is a subset of what the destination rectangle already covers.
// Diagonal steps also need both orthogonal neighbor positions clear, so the
// straight-line tween through the shared corner never crosses a wall tile.
func (p *Player) canStep(m *world.MapData, dir rl.Vector2) bool {
	h := p.halfTile()
	if p.blockedAt(m, rl.Vector2Add(p.logical, rl.Vector2Scale(dir, h))) {
		return false
	}
	if dir.X != 0 && dir.Y != 0 {
		if p.blockedAt(m, rl.Vector2Add(p.logical, rl.Vector2{X: dir.X * h, Y: 0})) {
			return false
		}
		if p.blockedAt(m, rl.Vector2Add(p.logical, rl.Vector2{X: 0, Y: dir.Y * h})) {
			return false
		}
	}
	return true
}

// blockedAt uses the same tile and object collision checks as the original
// continuous movement, just centered on the hypothetical position.
func (p *Player) blockedAt(m *world.MapData, logical rl.Vector2) bool {
	rect := rl.Rectangle{
		X:      logical.X - p.Size.X/2,
		Y:      logical.Y - p.Size.Y/2,
		Width:  p.Size.X,
		Height: p.Size.Y,
	}
	return m.CanCollidRect(rect) || m.CanCollideObjects(rect)
}

func (p *Player) applyLogical(logical rl.Vector2) {
	p.logical = logical
	p.Position = rl.Vector2{
		X: logical.X - p.Size.X/2,
		Y: logical.Y - p.Size.Y/2,
	}
}

func (p *Player) halfTile() float32 {
	return p.tileSize / 2
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

// Helper functions
func (p *Player) updateFacing(input rl.Vector2) {
	if input.X == 0 && input.Y == 0 {
		return
	}
	if math.Abs(float64(input.X)) == math.Abs(float64(input.Y)) {
		return
	}
	if math.Abs(float64(input.X)) > math.Abs(float64(input.Y)) {
		if input.X > 0 {
			p.Facing = rl.Vector2{X: 1, Y: 0}
		} else {
			p.Facing = rl.Vector2{X: -1, Y: 0}
		}
	} else {
		if input.Y > 0 {
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
	return p.logical
}

func (p *Player) Sit(seatRect rl.Rectangle) {
	p.savedPosition = p.Position
	p.applyLogical(rl.Vector2{
		X: seatRect.X + seatRect.Width/2,
		Y: seatRect.Y + seatRect.Height/2,
	})
	p.Sitting = true
	p.stepping = false
	p.progress = 0
}
func (p *Player) Stand() {
	p.applyLogical(rl.Vector2{
		X: p.savedPosition.X + p.Size.X/2,
		Y: p.savedPosition.Y + p.Size.Y/2,
	})
	p.Sitting = false
	p.stepping = false
	p.progress = 0
}
