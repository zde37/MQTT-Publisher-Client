package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Publisher struct {
	ID        string
	UserName  string
	Password  string
	ClientID  string
	BrokerURL string
	Topic     string
}

func NewPublisher(id, username, topic, clientID string) Publisher {
	return Publisher{
		ID:        id,
		BrokerURL: "", // replace with your broker url
		UserName:  username, // replace with your username
		Password:  "", // replace with your password
		ClientID:  clientID, // replace with your client ID
		Topic:     topic, // replace with your topic
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	temperaturePublisher := NewPublisher("Temperature", "", "topic/device/temperature", "") // replace with your publisher details
	speedPublisher := NewPublisher("Speed", "", "topic/device/speed", "")  // replace with your publisher details
	pressurePublisher := NewPublisher("Pressure", "", "topic/device/pressure", "")  // replace with your publisher details

	temperatureClient := newPublisherClient(temperaturePublisher)
	speedClient := newPublisherClient(speedPublisher)
	pressureClient := newPublisherClient(pressurePublisher)

	go func() {
		temperaturePublisher.publishMessage(temperatureClient, 3)
	}()
	
	go func() {
		speedPublisher.publishMessage(speedClient, 5)
	}()

	go func() {
		pressurePublisher.publishMessage(pressureClient, 7)
	}()
	
	<-c
}

func newPublisherClient(publisher Publisher) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(publisher.BrokerURL)
	opts.SetClientID(publisher.ClientID)
	opts.SetUsername(publisher.UserName)
	opts.SetPassword(publisher.Password)

	tlsConfig := newTLSConfig()
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		log.Printf("%s publisher connected", publisher.ID)
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("%s lost connection: %v\n", publisher.ID, err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("%s publisher failed to connect: %v", publisher.ID, token.Error())
	}

	return client
}

func (p Publisher) publishMessage(client mqtt.Client, duration time.Duration) {
	i := 0
	for range time.Tick(duration * time.Second) {
		text := fmt.Sprintf("%s is currently: %d", p.ID, i)
		token := client.Publish(p.Topic, 0, false, text)
		token.Wait()
		i++
	}
}

func newTLSConfig() *tls.Config {
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile("ssl_cert.crt") // replace with the path to your ssl certificate
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	certPool.AppendCertsFromPEM(ca)
	return &tls.Config{
		RootCAs: certPool,
	}
}
