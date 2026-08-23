package pkg

import (
	"context"
	"encoding/json"
	"home-monitor-backend/models"
	"home-monitor-backend/repositories"
	"home-monitor-backend/utils"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

var DeviceMessageQueue Queue
var deviceRepo repositories.DeviceRepository

const (
	DeviceTopic          = "req/device/#"
	DeviceResultTopic    = "res/device"
	DeviceOfflineTimeout = 5 * time.Minute
	StatusCheckInterval  = 5 * time.Minute
)

type ResultStatus int

const (
	ResultStatusSuccess ResultStatus = iota
	ResultStatusInvalidUUID
	ResultStatusInvalidToken
	ResultStatusInvalidDevice
	ResultStatusInvalidPayload
	ResultStatusFailedDatabase
)

func DeviceMainTask() {
	ctx := context.Background()

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

		if len(strings.Split(msg.Topic, "/")) < 4 {
			log.Printf("Invalid topic format: %s", msg.Topic)
			continue
		}

		deviceStrUUID := strings.Split(msg.Topic, "/")[2]
		if deviceStrUUID == "" {
			log.Printf("Empty device UUID in topic: %s", msg.Topic)
			continue
		}

		token := strings.Split(msg.Topic, "/")[3]
		if token == "" {
			log.Printf("Empty token in topic: %s", msg.Topic)
			continue
		}

		deviceValidateUUID, err := utils.ValidateDeviceJWT(token)
		if err != nil {
			log.Printf("Invalid token for device %s: %v", deviceStrUUID, err)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusInvalidToken,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		deviceUUID, err := uuid.Parse(deviceStrUUID)
		if err != nil {
			log.Printf("Error parsing device UUID %s: %v", deviceStrUUID, err)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusInvalidUUID,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		device, err := deviceRepo.DeviceFindByUUID(ctx, deviceUUID)
		if err != nil {
			log.Printf("Error finding device by UUID %s: %v", deviceStrUUID, err)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusInvalidUUID,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		if device.UUID != deviceValidateUUID {
			log.Printf("Token UUID does not match device UUID for device %s", deviceStrUUID)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusInvalidDevice,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		var payload models.DeviceMeasurementPayload
		err = json.Unmarshal(msg.Payload, &payload)
		if err != nil {
			log.Printf("Error unmarshaling payload for device %s: %v", deviceStrUUID, err)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusInvalidPayload,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		measurement := &models.DeviceMeasurement{
			DeviceID:    device.ID,
			Temperature: payload.Temperature,
			Humidity:    payload.Humidity,
		}

		err = deviceRepo.DeviceMeasurementCreate(ctx, measurement)
		if err != nil {
			log.Printf("Error creating measurement for device %s: %v", deviceUUID, err)

			response := models.DeviceMeasurementMqttResponse{
				Result: -ResultStatusFailedDatabase,
			}
			responseJson, _ := json.Marshal(response)
			err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
			if err != nil {
				log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
			}

			continue
		}

		err = deviceRepo.DeviceUpdateStatus(ctx, device.ID, true)
		if err != nil {
			log.Printf("Warning: Failed to update device status for device %s: %v", deviceUUID, err)
		}

		response := models.DeviceMeasurementMqttResponse{
			Result: ResultStatusSuccess,
		}
		responseJson, _ := json.Marshal(response)
		err = MqttPublish(DeviceResultTopic+"/"+deviceStrUUID+"/"+token, responseJson)
		if err != nil {
			log.Printf("Failed to publish MQTT response for device %s: %v", deviceStrUUID, err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func DeviceStatusChecker() {
	ctx := context.Background()

	ticker := time.NewTicker(StatusCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		devices, err := deviceRepo.DeviceList(ctx, 0, false)
		if err != nil {
			log.Printf("Error fetching devices for status check: %v", err)
			continue
		}

		for _, device := range devices {
			if !device.Status {
				continue
			}

			measurements, err := deviceRepo.DeviceMeasurementsByDeviceID(ctx, 1, device.ID, true, nil, nil)
			if err != nil {
				log.Printf("Error fetching measurements for device %s: %v", device.UUID, err)
				continue
			}

			if len(measurements) == 0 {
				continue
			}

			lastMeasurement := measurements[0]
			timeSinceLastMeasurement := time.Since(lastMeasurement.CreatedAt)

			if timeSinceLastMeasurement > DeviceOfflineTimeout {
				err = deviceRepo.DeviceUpdateStatus(ctx, device.ID, false)
				if err != nil {
					log.Printf("Error updating device status to offline for device %s: %v", device.UUID, err)
				} else {
					log.Printf("Device %s marked as offline (last seen: %v ago)", device.UUID, timeSinceLastMeasurement.Round(time.Second))
				}
			}
		}
	}
}

func DeviceInit(repo repositories.DeviceRepository) {
	deviceRepo = repo
	go DeviceMainTask()
	go DeviceStatusChecker()
}
