package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/eclipse/paho.mqtt.golang"
)

func runcmd(cmdStr, msg string) {
	if cmdStr == "" {
		return
	}
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Env = append(os.Environ(), "MQTT_MESSAGE="+msg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("Running", cmdStr)
	cmd.Run()
}

func main() {
	mqtt.ERROR = log.New(os.Stderr, "", 0)
	host := flag.String("host", os.Getenv("MQTT_HOST"), "mqtt hostname, defaults to what's in $MQTT_HOST")

	username := flag.String("user", os.Getenv("MQTT_USER"), "mqtt username, defaults to what's in $MQTT_USER")
	password := flag.String("password", "", "mqtt password, defaults to what's in $MQTT_PASS")
	topic := flag.String("t", "", "topicname")
	command := flag.String("cmd", "echo $MQTT_MESSAGE", "shell command to run. MQTT_MESSAGE will be set to the contents of the message")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()

	if *password == "" {
		*password = os.Getenv("MQTT_PASS")
	}

	if *host == "" {
		fmt.Fprint(os.Stderr, "ERROR: -host is required (or set $MQTT_HOST)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !strings.Contains(*host, ":") {
		*host = *host + ":1883"

	}

	opts := mqtt.NewClientOptions().
		AddBroker(*host).
		SetUsername(*username).
		SetPassword(*password).
		SetAutoReconnect(true).
		SetOnConnectHandler(func(c mqtt.Client) {
			token := c.Subscribe(*topic, 0, func(c mqtt.Client, m mqtt.Message) {
				runcmd(*command, string(m.Payload()))
			})

			if token.Wait() && token.Error() != nil {
				log.Fatal(token.Error())
			}
	})

	c := mqtt.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	defer func() {
		c.Disconnect(250)
	}()

	signalWait := make(chan os.Signal, 1)
	signal.Notify(signalWait, os.Interrupt, syscall.SIGTERM,  syscall.SIGINT)
	<-signalWait
}
