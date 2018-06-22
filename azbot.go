package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/mqtt"
)

func main() {
	deviceID := os.Getenv("MQTT_DEVICE_ID")
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	host := os.Getenv("MQTT_HOST")
	port := 8883
	brokerURL := fmt.Sprintf("tcps://%s:%d", host, port)

	mqttAdaptor := mqtt.NewAdaptorWithAuth(
		brokerURL,
		deviceID,
		username,
		password,
	)

	mqttAdaptor.SetServerCert("BaltimoreRootCertificate.cer")
	mqttAdaptor.SetUseSSL(true)

	err := mqttAdaptor.Connect()

	work := func() {
		data := []byte("hello")
		eventTopic := fmt.Sprintf("devices/%s/messages/events/", deviceID)
		incomingMessages := fmt.Sprintf("devices/%s/messages/devicebound/#", deviceID)

		mqttAdaptor.On(incomingMessages, func(msg mqtt.Message) {
			fmt.Printf("Receieved message: '%s' in topic '%s' with id '%d'\n", string(msg.Payload()[:]), msg.Topic(), msg.MessageID())
		})

		gobot.Every(5*time.Second, func() {
			fmt.Printf("Sending '%s' to IoT Hub topic %s\n", data, eventTopic)
			mqttAdaptor.Publish(eventTopic, data)
		})
	}

	robot := gobot.NewRobot("azbot",
		[]gobot.Connection{mqttAdaptor},
		work,
	)

	err = robot.Start()

	if err != nil {
		fmt.Println(err)
	}
}
