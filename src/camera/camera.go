package camera

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func New() *Controller {
	c := &Controller{
		View: rl.Camera2D{
			Zoom: 2.0,
		},
		FollowSpeed: 6.0,
	}

	c.UpdateViewport()
	return c
}

func (c *Controller) UpdateViewport() {
	c.View.Offset = rl.Vector2{
		X: float32(rl.GetScreenWidth()) / 2,
		Y: float32(rl.GetScreenHeight()) / 2,
	}
}

func (c *Controller) Follow(target rl.Vector2, dt float32) {
	alpha := float32(1 - math.Exp(float64(-dt*c.FollowSpeed)))

	c.View.Target.X += (target.X - c.View.Target.X) * alpha
	c.View.Target.Y += (target.Y - c.View.Target.Y) * alpha
}
