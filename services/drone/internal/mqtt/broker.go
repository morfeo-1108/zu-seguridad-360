package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type TelemetryPayload struct {
	DroneID      string    `json:"drone_id"`
	Timestamp    time.Time `json:"timestamp"`
	BatteryLevel int       `json:"battery_level"`
	GPSLat       float64   `json:"gps_lat"`
	GPSLng       float64   `json:"gps_lng"`
	GPSAlt       float64   `json:"gps_alt"`
	Speed        float64   `json:"speed_ms"`
	Status       string    `json:"status"`
}

type VisionEvent struct {
	DroneID    string                 `json:"drone_id"`
	CompanyID  string                 `json:"company_id"`
	EventType  string                 `json:"event_type"`
	Confidence float64                `json:"confidence"`
	BboxData   map[string]interface{} `json:"bbox_data"`
	PhotoURL   string                 `json:"photo_url"`
	Timestamp  time.Time              `json:"timestamp"`
}

type MQTTBroker struct {
	client      paho.Client
	telemetryCh chan TelemetryPayload
	visionCh    chan VisionEvent
}

func NewMQTTBroker(brokerURL string) *MQTTBroker {
	b := &MQTTBroker{
		telemetryCh: make(chan TelemetryPayload, 100),
		visionCh:    make(chan VisionEvent, 100),
	}
	opts := paho.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("zu-drone-service").
		SetAutoReconnect(true).
		SetOnConnectHandler(b.onConnect)
	b.client = paho.NewClient(opts)
	return b
}

func (b *MQTTBroker) Connect() error {
	token := b.client.Connect()
	token.Wait()
	return token.Error()
}

func (b *MQTTBroker) onConnect(c paho.Client) {
	log.Println("✅ MQTT EMQX connected")
	c.Subscribe("zu/drones/+/telemetry", 1, b.handleTelemetry)
	c.Subscribe("zu/drones/+/vision/events", 1, b.handleVisionEvent)
}

func (b *MQTTBroker) handleTelemetry(_ paho.Client, msg paho.Message) {
	var p TelemetryPayload
	// A malformed frame must be dropped, not passed on as zero telemetry
	// (0 % battery, 0 satellites) that would trigger failsafes.
	if err := json.Unmarshal(msg.Payload(), &p); err != nil {
		log.Printf("telemetry dropped (%s): %v", msg.Topic(), err)
		return
	}
	b.telemetryCh <- p
}

func (b *MQTTBroker) handleVisionEvent(_ paho.Client, msg paho.Message) {
	var e VisionEvent
	if err := json.Unmarshal(msg.Payload(), &e); err != nil {
		log.Printf("vision event dropped (%s): %v", msg.Topic(), err)
		return
	}
	b.visionCh <- e
}

func (b *MQTTBroker) SendMissionCommand(droneID string, command map[string]interface{}) error {
	topic := fmt.Sprintf("zu/drones/%s/commands", droneID)
	payload, _ := json.Marshal(command)
	token := b.client.Publish(topic, 1, false, payload)
	token.Wait()
	return token.Error()
}

func (b *MQTTBroker) TelemetryCh() <-chan TelemetryPayload { return b.telemetryCh }
func (b *MQTTBroker) VisionCh() <-chan VisionEvent         { return b.visionCh }

// Publish envía un payload crudo a un topic (usado por el piloto autónomo DJI).
func (b *MQTTBroker) Publish(topic string, payload []byte) error {
	if b.client == nil {
		return nil // modo dev sin broker
	}
	token := b.client.Publish(topic, 1, false, payload)
	token.Wait()
	return token.Error()
}
