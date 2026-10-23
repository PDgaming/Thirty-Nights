package player

// Constants
var animations = Animations{
	Idle: Animation{
		FrameCount:    1,
		FrameDuration: 0.15,
		Loop:          true,
	},
	Walking: Animation{
		FrameCount:    2,
		FrameDuration: 0.15,
		Loop:          true,
	},
}

func NewAnimator() *Animator {
	a := &Animator{
		State:      Idle,
		Frame:      0,
		Elapsed:    0,
		Animations: animations,
	}
	return a
}

func (a *Animator) ChangeState(state AnimationState) {
	if a.State == state {
		return
	}

	a.State = state
	a.Frame = 0
	a.Elapsed = 0
}

func (a *Animator) Update(dt float32) {
	if a.State == Idle {
		a.Frame = 0
		return
	}

	a.Elapsed += dt
	if a.Elapsed >= a.Animations[a.State].FrameDuration {
		a.Frame += 1
		a.Elapsed -= a.Animations[a.State].FrameDuration
	}

	if a.Animations[a.State].Loop && a.Frame >= a.Animations[a.State].FrameCount {
		a.Frame = 0
	}
}
