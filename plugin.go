package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/gotify/plugin-api"
)

// GetGotifyPluginInfo provides metadata about the Gotify plugin.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "github.com/FaintGhost/gotify-webhook-mqtt",
		Version:     "0.0.1",
		Author:      "FaintGhost",
		Website:     "https://github.com/FaintGhost/gotify-webhook-mqtt",
		Description: "A Gotify plugin for MQTT message forwarding with customizable settings.",
		License:     "MIT",
		Name:        "gotify/mqtt",
	}
}

// MQTTConfig holds the configuration for the MQTT connection.
type MQTTConfig struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	QoS       byte   `json:"qos"`
	Topic     string `json:"topic"`
	KeepAlive int    `json:"keepAlive"`
	Retain    bool   `json:"retain"`
}

// MQTTPlugin defines the structure for the Gotify plugin with MQTT support.
type MQTTPlugin struct {
	config     *MQTTConfig
	msgHandler plugin.MessageHandler
	mqttClient mqtt.Client
	basePath   string
	lastError  error

	// For message reliability
	messageQueue []plugin.Message
	mu           sync.Mutex
}

// NewGotifyPluginInstance initializes a new instance of the MQTTPlugin.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &MQTTPlugin{}
}

// Enable establishes a connection to the MQTT broker using the provided config.
func (p *MQTTPlugin) Enable() error {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s:%d", p.config.IP, p.config.Port))
	opts.SetUsername(p.config.Username)
	opts.SetPassword(p.config.Password)
	opts.SetKeepAlive(time.Duration(p.config.KeepAlive) * time.Second) // Set KeepAlive

	p.mqttClient = mqtt.NewClient(opts)
	if token := p.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		p.lastError = fmt.Errorf("Failed to connect to MQTT Broker: %v", token.Error())
		log.Println(p.lastError)
		return p.lastError
	}

	// Resend messages that might have been queued during a disconnect
	p.mu.Lock()
	for _, msg := range p.messageQueue {
		p.forwardToMQTT(&msg)
	}
	p.messageQueue = nil
	p.mu.Unlock()

	log.Println("Successfully connected to MQTT Broker")
	return nil
}

// Disable disconnects from the MQTT broker.
func (p *MQTTPlugin) Disable() error {
	p.mqttClient.Disconnect(250)
	return nil
}

// SetMessageHandler sets the message handler for the plugin.
func (p *MQTTPlugin) SetMessageHandler(h plugin.MessageHandler) {
	p.msgHandler = h
}

// RegisterWebhook registers the endpoint for receiving messages and forwarding them to MQTT.
func (p *MQTTPlugin) RegisterWebhook(basePath string, mux *gin.RouterGroup) {
	p.basePath = basePath
	mux.POST("/message", func(c *gin.Context) {
		var msg plugin.Message
		err := c.BindJSON(&msg) // Added error handling
		if err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse message"})
			return
		}

		// Send the message to Gotify first
		err = p.msgHandler.SendMessage(msg)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to send message to Gotify"})
			return
		}

		// Then forward the message to MQTT or queue it if not connected
		p.mu.Lock()
		if p.mqttClient != nil && p.mqttClient.IsConnected() {
			p.forwardToMQTT(&msg)
		} else {
			p.messageQueue = append(p.messageQueue, msg)
		}
		p.mu.Unlock()

		c.JSON(200, gin.H{"status": "Message forwarded to MQTT and Gotify"})
	})
}

// forwardToMQTT sends the message to the MQTT broker.
func (p *MQTTPlugin) forwardToMQTT(msg *plugin.Message) {
	token := p.mqttClient.Publish(p.config.Topic, p.config.QoS, p.config.Retain, msg.Message)
	token.Wait()
}

// GetDisplay provides a display message for the plugin's UI.
func (p *MQTTPlugin) GetDisplay(location *url.URL) string {
	if p.lastError != nil {
		return fmt.Sprintf("Error: %s\n\nSend messages to %s%smessage to forward them to MQTT.", p.lastError, location.String(), p.basePath)
	}
	return fmt.Sprintf("Send messages to %s%smessage to forward them to MQTT.", location.String(), p.basePath)
}

// DefaultConfig provides the default configuration for the MQTT plugin.
func (p *MQTTPlugin) DefaultConfig() interface{} {
	return &MQTTConfig{
		IP:        "127.0.0.1",
		Port:      1883,
		QoS:       0,
		Topic:     "gotify/messages",
		KeepAlive: 60, // default to 60 seconds
		Retain:    false,
	}
}

// ValidateAndSetConfig validates the provided configuration and sets it for the plugin.
func (p *MQTTPlugin) ValidateAndSetConfig(conf interface{}) error {
	config := conf.(*MQTTConfig)

	// Validate IP
	if net.ParseIP(config.IP) == nil {
		return fmt.Errorf("invalid IP address: %s", config.IP)
	}

	// Validate port
	if config.Port < 1 || config.Port > 65535 {
		return fmt.Errorf("port out of range: %d", config.Port)
	}

	// Validate KeepAlive
	if config.KeepAlive < 0 {
		return fmt.Errorf("invalid KeepAlive value: %d", config.KeepAlive)
	}

	p.config = config
	return nil
}

func main() {
	panic("this should be built as go plugin")
}
