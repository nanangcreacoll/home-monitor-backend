package pkg

import (
	"log"
	"os"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var mqttClient MQTT.Client

var broker = ""
var username = ""
var password = ""
var clientID = ""

type MqttMessage struct {
	Topic   string
	Payload []byte
}

var messagePubHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	log.Printf("Received from topic: %s with payload: %s\n", msg.Topic(), string(msg.Payload()))

	topic := msg.Topic()
	if strings.HasPrefix(topic, DeviceTopic[:len(DeviceTopic)-1]) {
		DeviceMessageQueue.Push(MqttMessage{
			Topic:   msg.Topic(),
			Payload: msg.Payload(),
		})
	}
}

var connectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	log.Printf("Connected to MQTT broker at %s\n", broker)
}

var connectLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	log.Printf("Connection lost: %v", err)

	log.Printf("Attempting to reconnect to MQTT broker at %s\n", broker)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Reconnection failed: %v", token.Error())
	} else {
		log.Printf("Reconnected to MQTT broker at %s\n", broker)
	}
}

func MqttInit() error {
	broker = os.Getenv("MQTT_BROKER")
	username = os.Getenv("MQTT_USER")
	password = os.Getenv("MQTT_PASSWORD")
	clientID = os.Getenv("MQTT_CLIENT_ID")

	if broker == "" {
		log.Fatal("MQTT_BROKER environment variable is not set")
		return nil
	}

	if username == "" {
		log.Fatal("MQTT_USER environment variable is not set")
		return nil
	}

	if password == "" {
		log.Fatal("MQTT_PASSWORD environment variable is not set")
		return nil
	}

	if clientID == "" {
		log.Fatal("MQTT_CLIENT_ID environment variable is not set")
		return nil
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	mqttClient = MQTT.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("MQTT connection error: ", token.Error())
		return token.Error()
	}

	return nil
}

func MqttPublish(topic string, payload []byte) error {
	token := mqttClient.Publish(topic, 0, false, payload)
	token.Wait()
	log.Printf("Published message to topic %s: %s\n", topic, string(payload))
	return token.Error()
}

func MqttSubscribe(topic string) error {
	token := mqttClient.Subscribe(topic, 0, nil)
	token.Wait()
	log.Printf("Subscribed to topic: %s\n", topic)
	return token.Error()
}

func MqttUnsubscribe(topic string) error {
	token := mqttClient.Unsubscribe(topic)
	token.Wait()
	log.Printf("Unsubscribed from topic: %s\n", topic)
	return token.Error()
}
