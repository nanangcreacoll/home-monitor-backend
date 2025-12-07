package pkg

import (
	"encoding/json"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var DeviceMessageQueue Queue
var deviceRepo repositories.DeviceRepository

const (
	DeviceTopic = "req/device/#"
)

func DeviceMainTask() {
	err := MqttSubscribe(DeviceTopic)
	if err != nil {
		log.Fatalf("Failed to subscribe to MQTT topic: %v", err)
	}

	DeviceMessageQueue = NewQueue()
	for {
		if DeviceMessageQueue.Empty() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		msg := DeviceMessageQueue.Pop().(MqttMessage)

		if len(strings.Split(msg.Topic, "/")) < 3 {
			log.Printf("Invalid topic format: %s", msg.Topic)
			continue
		}

		deviceUUID := strings.Split(msg.Topic, "/")[2]
		if deviceUUID == "" {
			log.Printf("Empty device UUID in topic: %s", msg.Topic)
			continue
		}

		uuidParsed, err := uuid.Parse(deviceUUID)
		if err != nil {
			log.Printf("Error parsing device UUID %s: %v", deviceUUID, err)
			continue
		}

		device, err := deviceRepo.DeviceFindByUUID(uuidParsed)
		if err != nil {
			log.Printf("Error finding device by UUID %s: %v", deviceUUID, err)
			continue
		}

		var payload models.DeviceMeasurementPayload
		err = json.Unmarshal(msg.Payload, &payload)
		if err != nil {
			log.Printf("Error unmarshaling payload for device %s: %v", deviceUUID, err)
			continue
		}

		measurement := &models.DeviceMeasurement{
			DeviceID:    device.ID,
			Temperature: payload.Temperature,
			Humidity:    payload.Humidity,
		}

		err = deviceRepo.DeviceMeasurementCreate(measurement)
		if err != nil {
			log.Printf("Error creating measurement for device %s: %v", deviceUUID, err)
			continue
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func DeviceInit(repo repositories.DeviceRepository) {
	deviceRepo = repo
	go DeviceMainTask()
}
