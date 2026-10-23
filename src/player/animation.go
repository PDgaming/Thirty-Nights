package player

func NewAnimator() *Animator {
	a := &Animator{
		State:         Idle,
		Frame:         0,
		Elapsed:       0,
		FrameDuration: 0.15,
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
	if a.Elapsed >= a.FrameDuration {
		a.Frame += 1
		a.Elapsed -= a.FrameDuration
	}
}
