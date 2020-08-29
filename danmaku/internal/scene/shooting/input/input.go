package input

import (
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"

	"github.com/yohamta/godanmaku/danmaku/internal/scene/shooting/input/firebutton"
	"github.com/yohamta/godanmaku/danmaku/internal/touch"
	"github.com/yohamta/godanmaku/danmaku/internal/ui/joystick"
)

// Input represents the state of user's input
type Input struct {
	Horizontal   float64
	Vertical     float64
	Fire         bool
	prevTickTime time.Time
	joystick     *joystick.Joystick
	fireButton   *firebutton.FireButton
}

// New creates new Input
func New() *Input {
	input := &Input{}
	input.prevTickTime = time.Now()
	input.joystick = joystick.New()
	input.fireButton = firebutton.New()
	return input
}

// Update updates the input state
func (input *Input) Update() {
	if touch.IsTouchPrimaryInput() {
		input.readTouchInput()
		input.joystick.Update()
		input.fireButton.Update()
	} else {
		input.readKeyboardInput()
	}
}

// Draw draws input UI
func (input *Input) Draw(screen *ebiten.Image) {
	if input.joystick.IsReadingTouch() {
		input.joystick.Draw(screen)
	}
	input.fireButton.Draw(screen)
}

func (input *Input) readTouchInput() {
	justPressedTouchIds := inpututil.JustPressedTouchIDs()
	jStick := input.joystick
	fButton := input.fireButton

	if justPressedTouchIds != nil {
		for i := 0; i < len(justPressedTouchIds); i++ {
			touchID := justPressedTouchIds[i]
			if fButton.HandleTouch(touchID) {
				continue
			}
			if jStick.IsReadingTouch() == false {
				jStick.StartReadingTouch(justPressedTouchIds[0])
			}
		}
	}

	if fButton.IsPressing() {
		fButton.CheckIsTouchRelased()
	}

	if jStick.IsReadingTouch() {
		if inpututil.IsTouchJustReleased(jStick.GetTouchID()) {
			jStick.EndReadingTouch()
			input.Horizontal = 0
			input.Vertical = 0
		} else {
			input.Horizontal, input.Vertical = jStick.ReadInput()
		}
	}
}

func (input *Input) readKeyboardInput() {
	if time.Since(input.prevTickTime).Milliseconds() < 50 {
		// Adjust sensitivity for keyboard input
		return
	}
	input.prevTickTime = time.Now()

	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		input.Vertical = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		input.Vertical = -1
	} else {
		input.Vertical = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		input.Horizontal = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		input.Horizontal = -1
	} else {
		input.Horizontal = 0
	}

	input.Fire = ebiten.IsKeyPressed(ebiten.KeySpace)
}