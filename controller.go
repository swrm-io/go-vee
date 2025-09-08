package govee

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type Controller struct {
	logger  *slog.Logger
	devices []*Device
	ctx     context.Context
	command chan Message
}

func NewController(logger *slog.Logger) *Controller {

	return &Controller{
		devices: []*Device{},
		logger:  logger,
		ctx:     context.Background(),
		command: make(chan Message),
	}
}

// Start initializes the controller, begins listening for device messages,
// and starts periodic scanning for devices (every 60 seconds).
func (c *Controller) Start() error {
	c.logger.Info("Starting Govee Controller")
	addr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:4002")
	if err != nil {
		c.logger.Error("Failed to resolve UDP address", "error", err)
		return err
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		c.logger.Error("Failed to listen on multicast UDP", "error", err)
		return err
	}
	defer conn.Close()

	conn.SetReadBuffer(8192)

	// ToDo: make this a waitgroup and have a shutdown function
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return

			default:
				buffer := make([]byte, 8192)
				n, src, err := conn.ReadFromUDP(buffer)
				if err != nil {
					c.logger.Error("Error reading from UDP", "error", err)
					continue
				}

				srcAddr := src.IP.String()
				device, err := c.DeviceByIP(srcAddr)

				// New device discovered, register it and start its handler.
				if err != nil {
					c.logger.Debug("Discovered new device", "ip", srcAddr)

					deviceLogger := c.logger.With("device_ip", srcAddr)
					newDevice := Device{
						ip:       srcAddr,
						logger:   deviceLogger,
						ctx:      c.ctx,
						command:  c.command,
						response: make(chan Message),
					}
					go newDevice.handler()
					c.devices = append(c.devices, &newDevice)
					device = &newDevice
				}

				// Parse incoming message
				var request wrapper
				err = json.Unmarshal(buffer[:n], &request)
				if err != nil {
					c.logger.Error("Invalid API Request", "error", err)
					continue
				}

				// Handle incoming command and dispatch to device handler
				switch request.MSG.CMD {
				case "scan":
					c.logger.Debug("Received scan response", "from", srcAddr)
					msg := scanResponse{}
					err = json.Unmarshal(request.MSG.Data, &msg)
					if err != nil {
						c.logger.Error("Invalid scan response", "error", err)
						continue
					}

					device.response <- Message{IP: srcAddr, Payload: msg}

				case "devStatus":
					c.logger.Debug("Received device status", "from", srcAddr)
					msg := devStatusResponse{}
					err = json.Unmarshal(request.MSG.Data, &msg)
					if err != nil {
						c.logger.Error("Invalid device status response", "error", err)
						continue
					}

					device.response <- Message{IP: srcAddr, Payload: msg}

				default:
					c.logger.Warn("Unknown command received", "cmd", request.MSG.CMD)
				}
			}
		}
	}()

	go func() {
		for cmd := range c.command {
			data, err := json.Marshal(cmd.Payload)
			if err != nil {
				c.logger.Error("Failed to marshal command", "error", err)
				conn.Close()
				continue
			}

			var target string
			if cmd.IP == "239.255.255.250" {
				target = fmt.Sprintf("%s:4001", cmd.IP)
			} else {
				target = fmt.Sprintf("%s:4003", cmd.IP)
			}

			addr, err := net.ResolveUDPAddr("udp4", target)
			if err != nil {
				c.logger.Error("Failed to resolve device address", "error", err)
				conn.Close()
				continue
			}

			conn, err := net.DialUDP("udp4", nil, addr)
			if err != nil {
				c.logger.Error("Failed to dial device address", "error", err)
				conn.Close()
				continue
			}

			_, err = conn.Write(data)
			if err != nil {
				c.logger.Error("Failed to send command", "error", err)
				conn.Close()
				continue
			}

			conn.Close()
		}
	}()

	go func() error {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		scan, err := newAPIRequest("scan", scanRequest{AccountTopic: "reserve"})
		if err != nil {
			c.logger.Error("Failed to create scan request", "error", err)
			return err
		}
		msg := Message{"239.255.255.250", scan}

		// send immediate scan on startup
		c.command <- msg

		for {
			select {
			case <-c.ctx.Done():
				return nil
			case <-ticker.C:
				c.logger.Debug("Sending periodic scan request")
				c.command <- msg
			}
		}
	}()

	<-c.ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the controller.
func (c *Controller) Shutdown() error {
	c.logger.Info("Shutting down Govee Controller")
	c.ctx.Done()
	return nil
}

// Devices returns a slice of all managed devices.
func (c *Controller) Devices() []*Device {
	return c.devices
}

// DeviceByIP returns a pointer to a device by its IP address, or nil if not found.
func (c *Controller) DeviceByIP(ip string) (*Device, error) {
	for _, device := range c.devices {
		if device.ip == ip {
			return device, nil
		}
	}
	return nil, ErrNoDeviceFound
}

// DeviceByID returns a pointer to a device by its DeviceID, or nil if not found.
func (c *Controller) DeviceByID(id string) (*Device, error) {
	for _, device := range c.devices {
		if device.deviceID == id {
			return device, nil
		}
	}
	return nil, ErrNoDeviceFound
}
