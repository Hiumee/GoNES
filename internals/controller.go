package internals

// https://wiki.nesdev.org/w/index.php?title=Input_devices

type Controller struct {
	state [8]bool
	pool  uint8
	count uint8
}

func (controller *Controller) SetInput(state [8]bool) {
	controller.state = state
}

func (controller *Controller) ReadState() uint8 {
	var state uint8 = 0
	if controller.count < 8 && controller.state[controller.count] {
		state = 1
	}
	controller.count++
	if controller.pool == 1 {
		controller.count = 0
	}
	return state
}

func (controller *Controller) WriteState(value uint8) {
	controller.pool = value & 0x01
	if value == 1 {
		controller.count = 0
	}
}
