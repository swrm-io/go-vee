package govee

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Device represents a Govee device with its properties and current state.
type Device struct {
	seen time.Time

	ip              string
	deviceID        string
	sku             string
	bleVersionHard  Version
	bleVersionSoft  Version
	wifiVersionHard Version
	wifiVersionSoft Version

	state       State
	brightness  Brightness
	color       Color
	colorKelvin ColorKelvin

	logger   *slog.Logger
	ctx      context.Context
	command  chan Message
	response chan Message
}

// ToDo: make it honory ctx.Done()
func (d *Device) handler() {
	for resp := range d.response {
		switch payload := resp.Payload.(type) {
		case scanResponse:
			d.logger.Info("Discovered device", "ip", payload.IP, "deviceID", payload.DeviceID, "sku", payload.SKU)
			d.ip = payload.IP
			d.deviceID = payload.DeviceID
			d.sku = payload.SKU
			d.bleVersionHard = payload.BleVersionHard
			d.bleVersionSoft = payload.BleVersionSoft
			d.wifiVersionHard = payload.WifiVersionHard
			d.wifiVersionSoft = payload.WifiVersionSoft
			d.seen = time.Now()

		case devStatusResponse:
			d.logger.Info("Device status update", "onOff", payload.OnOff, "brightness", payload.Brightness, "color", payload.Color, "colorKelvin", payload.ColorKelvin)
			d.state = payload.OnOff
			d.brightness = payload.Brightness
			d.color = payload.Color
			d.colorKelvin = payload.ColorKelvin
			d.seen = time.Now()
		default:
			d.logger.Warn("Unknown command type", "type", fmt.Sprintf("%T", resp))
		}
	}
}

// String returns a string representation of the device.
func (d Device) String() string {
	var sku = "unknown"
	if d.sku != "" {
		sku = d.sku
	}

	var deviceID = "unknown"
	if d.deviceID != "" {
		deviceID = d.deviceID
	}
	return fmt.Sprintf("%s: %s (%s)", sku, d.ip, deviceID)
}

// Active returns true if the device has been seen in the last 5 minutes.
func (d Device) Active() bool {
	return time.Since(d.seen) < 5*time.Minute
}

// Accessor methods for Device properties.
func (d Device) IP() string               { return d.ip }
func (d Device) DeviceID() string         { return d.deviceID }
func (d Device) SKU() string              { return d.sku }
func (d Device) BleVersionHard() Version  { return d.bleVersionHard }
func (d Device) BleVersionSoft() Version  { return d.bleVersionSoft }
func (d Device) WifiVersionHard() Version { return d.wifiVersionHard }
func (d Device) WifiVersionSoft() Version { return d.wifiVersionSoft }
func (d Device) State() State             { return d.state }
func (d Device) Brightness() Brightness   { return d.brightness }
func (d Device) Color() Color             { return d.color }
func (d Device) ColorKelvin() ColorKelvin { return d.colorKelvin }

// TurnOn turns the device on.
func (d Device) TurnOn() {
	d.logger.Debug("Sending Turn On command")
	cmd := onOffRequest{Value: 1}

	wrapper, _ := newAPIRequest("turn", cmd)
	d.command <- Message{IP: d.ip, Payload: wrapper}
}

// TurnOff turns the device off.
func (d Device) TurnOff() {
	d.logger.Debug("Sending Turn Off command")
	cmd := onOffRequest{Value: 0}

	wrapper, _ := newAPIRequest("turn", cmd)
	d.command <- Message{IP: d.ip, Payload: wrapper}
}

// Toggle toggles the device state.
func (d Device) Toggle() {
	d.logger.Debug("Toggling device state")
	if d.state == 1 {
		d.TurnOff()
	} else {
		d.TurnOn()
	}
}

// SetBrightness sets the brightness of the device.
func (d Device) SetBrightness(brightness Brightness) {
	d.logger.Debug("Setting brightness", "brightness", brightness)
	cmd := brightnessRequest{Value: brightness}

	wrapper, _ := newAPIRequest("brightness", cmd)
	d.command <- Message{IP: d.ip, Payload: wrapper}
}

// SetColor sets the color of the device.
func (d Device) SetColor(color Color) {
	d.logger.Debug("Setting color", "color", color)
	cmd := colorRequest{Color: color, Kelvin: 0}

	wrapper, _ := newAPIRequest("colorwc", cmd)
	d.command <- Message{IP: d.ip, Payload: wrapper}
}

// SetColorKelvin sets the color temperature of the device.
func (d Device) SetColorKelvin(colorKelvin ColorKelvin) {
	d.logger.Debug("Setting color temperature", "colorKelvin", colorKelvin)
	cmd := colorRequest{Color: Color{}, Kelvin: colorKelvin}

	wrapper, _ := newAPIRequest("colorKelvin", cmd)
	d.command <- Message{IP: d.ip, Payload: wrapper}
}
