package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/mqtt"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

	shutdown := make(chan bool, 1)

	work := func() {
		eventTopic := fmt.Sprintf("devices/%s/messages/events/", deviceID)
		incomingMessages := fmt.Sprintf("devices/%s/messages/devicebound/#", deviceID)

		mqttAdaptor.On(incomingMessages, func(msg mqtt.Message) {
			payload := string(msg.Payload()[:])
			fmt.Printf("Received message: '%s' in topic '%s' with id '%d'\n", payload, msg.Topic(), msg.MessageID())

			if payload == "shutdown" {
				fmt.Println("device received shutdown command")
				shutdown <- true
			}
		})

		gobot.Every(5*time.Second, func() {
			data := intToBytes(getTemperature())
			fmt.Printf("Sending '%s' to IoT Hub topic %s\n", data, eventTopic)
			mqttAdaptor.Publish(eventTopic, data)
		})
	}

	robot := gobot.NewRobot("azbot",
		[]gobot.Connection{mqttAdaptor},
		work,
	)

	err = robot.Start(false)

	if err != nil {
		fmt.Println(err)
	}

	<-shutdown
	robot.Stop()
}

func intToBytes(i int) []byte {
	return []byte(strconv.Itoa(i))
}

func getTemperature() int {
	return randInRange(10, 40)
}

func randInRange(min, max int) int {
	return rand.Intn(max-min) + min
}
