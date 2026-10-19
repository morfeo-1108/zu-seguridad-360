package flight

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// MQTTPublisher abstrae el broker (EMQX) hacia el dron.
type MQTTPublisher interface {
	Publish(topic string, payload []byte) error
}

// DJICommander implementa DroneCommander usando el patrón de comandos de la
// DJI Cloud API: se publican comandos MQTT al thing del dron y este los ejecuta
// con su propio control de vuelo + APAS (evasión de obstáculos onboard).
//
// IMPORTANTE: la evasión de obstáculos en tiempo real la hace el HARDWARE del
// dron (sensores APAS de DJI), no este software. El software planifica y decide;
// el dron ejecuta con sus sistemas de seguridad redundantes. Así debe ser: la
// seguridad crítica no depende solo de la red.
type DJICommander struct {
	droneSN string
	pub     MQTTPublisher
}

func NewDJICommander(droneSN string, pub MQTTPublisher) *DJICommander {
	return &DJICommander{droneSN: droneSN, pub: pub}
}

func (d *DJICommander) topic() string {
	return fmt.Sprintf("thing/product/%s/services", d.droneSN)
}

func (d *DJICommander) command(method string, data map[string]interface{}) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"tid":       fmt.Sprintf("%d", time.Now().UnixNano()),
		"bid":       d.droneSN,
		"timestamp": time.Now().UnixMilli(),
		"method":    method,
		"data":      data,
	})
	return d.pub.Publish(d.topic(), payload)
}

func (d *DJICommander) Arm(ctx context.Context) error {
	return d.command("flighttask_prepare", map[string]interface{}{"action": "arm"})
}
func (d *DJICommander) Takeoff(ctx context.Context, alt float64) error {
	return d.command("takeoff_to_point", map[string]interface{}{"target_altitude": alt})
}
func (d *DJICommander) GotoWaypoint(ctx context.Context, wp Waypoint) error {
	return d.command("fly_to_point", map[string]interface{}{
		"latitude": wp.Point.Lat, "longitude": wp.Point.Lng,
		"altitude": wp.Point.Alt, "max_speed": wp.SpeedMS,
	})
}
func (d *DJICommander) SetGimbal(ctx context.Context, pitch float64) error {
	return d.command("gimbal_set", map[string]interface{}{"pitch": pitch})
}
func (d *DJICommander) CapturePhoto(ctx context.Context) error {
	return d.command("camera_photo_take", map[string]interface{}{})
}
func (d *DJICommander) ReturnToHome(ctx context.Context) error {
	return d.command("return_home", map[string]interface{}{})
}
func (d *DJICommander) Land(ctx context.Context) error {
	return d.command("land", map[string]interface{}{})
}
func (d *DJICommander) EmergencyLand(ctx context.Context) error {
	return d.command("emergency_land", map[string]interface{}{"force": true})
}
func (d *DJICommander) Hover(ctx context.Context) error {
	return d.command("hover", map[string]interface{}{})
}
