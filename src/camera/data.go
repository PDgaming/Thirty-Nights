package camera

import rl "github.com/gen2brain/raylib-go/raylib"

type Controller struct {
	View        rl.Camera2D
	FollowSpeed float32
}
