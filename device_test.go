package govee

func ExampleDevice_TurnOn() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()
	device, _ := controller.DeviceByIP("192.168.1.100")
	device.TurnOn()
}
func ExampleDevice_TurnOff() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()

	device, _ := controller.DeviceByIP("192.168.1.100")
	device.TurnOff()
}

func ExampleDevice_Toggle() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()

	device, _ := controller.DeviceByIP("192.168.1.100")
	device.Toggle()
}

func ExampleDevice_SetBrightness() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()

	device, _ := controller.DeviceByIP("192.168.1.100")
	brightness := NewBrightness(75)
	device.SetBrightness(brightness)
}

func ExampleDevice_SetColor() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()

	device, _ := controller.DeviceByIP("192.168.1.100")
	color := NewColor(255, 0, 0)
	device.SetColor(color)
}

func ExampleDevice_SetColorKelvin() {
	controller := NewController(nil)
	go controller.Start()
	defer controller.Shutdown()

	device, _ := controller.DeviceByIP("192.168.1.100")
	k := NewColorKelvin(3500)
	device.SetColorKelvin(k)
}
