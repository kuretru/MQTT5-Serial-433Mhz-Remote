package mqtt

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"mqtt5-serial-433mhz-remote/internal"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func InstallMQTTClient(rootCtx context.Context, done func(), config *internal.AppConfig, commandChan chan string, stateChan chan string) {
	defer done()
	ctx, cancel := context.WithCancel(context.Background())

	brokerUrl, err := url.Parse(config.MQTT.URL)
	if err != nil {
		log.Fatalf("[AutoPaho] Parse server URL failed %s\n", config.MQTT.URL)
	}

	handler := messageHandler{commandChan}

	clientConfig := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{brokerUrl},
		KeepAlive:         uint16(config.MQTT.KeepAlive),
		ConnectRetryDelay: time.Duration(config.MQTT.ConnectRetryDelay) * time.Second,
		ConnectTimeout:    time.Duration(config.MQTT.ConnectTimeout) * time.Second,

		OnConnectionUp: func(connectionManager *autopaho.ConnectionManager, connack *paho.Connack) {
			_, err := connectionManager.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					config.MQTT.CommandTopic: {QoS: 0},
				},
			})
			if err != nil {
				log.Printf("[AutoPaho] Failed to subscribe (%s)\n", err)
				return
			}
			log.Printf("[AutoPaho] Subscribed %s QoS=0\n", config.MQTT.CommandTopic)

			sendOnlineMessage(ctx, config, connectionManager)
		},
		OnConnectError: func(err error) {
			log.Printf("[AutoPaho] Connect error, %s\n", err)
		},

		PahoErrors: logger{"Paho-error"},

		ClientConfig: paho.ClientConfig{
			ClientID: config.MQTT.ClientId,
			Router: paho.NewSingleHandlerRouter(func(data *paho.Publish) {
				handler.handle(data)
			}),
			OnServerDisconnect: func(disconnect *paho.Disconnect) {
				if disconnect.Properties != nil {
					log.Printf("[Paho] Server requested disconnect: %s\n", disconnect.Properties.ReasonString)
				} else {
					log.Printf("[Paho] Server requested disconnect, reason code: %d\n", disconnect.ReasonCode)
				}
			},
			OnClientError: func(err error) {
				log.Printf("[Paho] Server requested disconnect: %s\n", err)
			},
		},
	}
	clientConfig.SetWillMessage(config.MQTT.AvailabilityTopic, []byte("offline"), 1, true)
	if config.MQTT.Username != "" && config.MQTT.Password != "" {
		clientConfig.SetUsernamePassword(config.MQTT.Username, []byte(config.MQTT.Password))
	}
	if config.Debug {
		clientConfig.Debug = logger{"AutoPaho-debug"}
		clientConfig.PahoDebug = logger{"Paho-debug"}
	}

	connection, err := autopaho.NewConnection(ctx, clientConfig)
	if err != nil {
		panic(err)
	}
	log.Printf("[AutoPaho] Connected to server %s\n", brokerUrl)

loop:
	for {
		select {
		case <-rootCtx.Done():
			sendOfflineMessage(ctx, config, connection)
			log.Println("[Paho] Exiting")
			cancel()
			break loop
		case state := <-stateChan:
			err = connection.AwaitConnection(ctx)
			if err != nil {
				fmt.Printf("[AutoPaho] Connection error, wait next topic %s\n", err)
				continue
			}

			log.Printf("[Paho] Sending state topic: %s\n", state)
			_, err := connection.Publish(ctx, &paho.Publish{
				QoS:     0,
				Retain:  true,
				Topic:   config.MQTT.StateTopic,
				Payload: []byte(state),
			})
			if err != nil {
				log.Printf("[Paho] Send state topic failed: %s\n", err)
			}
		}
	}
}

func sendOnlineMessage(ctx context.Context, config *internal.AppConfig, connection *autopaho.ConnectionManager) {
await:
	err := connection.AwaitConnection(ctx)
	if err != nil && strings.Contains(err.Error(), "connection manager") {
		log.Println("Connection lost, waiting...")
		goto await
	}

	_, err = connection.Publish(ctx, &paho.Publish{
		QoS:     1,
		Retain:  true,
		Topic:   config.MQTT.AvailabilityTopic,
		Payload: []byte("online"),
	})
	if err != nil {
		log.Printf("[Paho] Send \"online\" message failed: %s\n", err)
	} else {
		log.Println("[AutoPaho] Sent \"online\" message")
	}
}

func sendOfflineMessage(ctx context.Context, config *internal.AppConfig, connection *autopaho.ConnectionManager) {
	err := connection.AwaitConnection(ctx)
	if err != nil && strings.Contains(err.Error(), "connection manager") {
		log.Println("Connection lost, not send \"offline\" message, but we have will message")
		return
	}

	_, err = connection.Publish(ctx, &paho.Publish{
		QoS:     0,
		Retain:  true,
		Topic:   config.MQTT.AvailabilityTopic,
		Payload: []byte("offline"),
	})
	if err != nil {
		log.Printf("[Paho] Send \"offline\" message failed: %s\n", err)
	} else {
		log.Println("[AutoPaho] Sent \"offline\" message")
	}
}

type messageHandler struct {
	commandChannel chan string
}

func (m *messageHandler) handle(data *paho.Publish) {
	payload := string(data.Payload[:])
	log.Printf("[Paho] Received command topic: %s\n", payload)
	m.commandChannel <- payload
}

// logger implements the paho.Logger interface
type logger struct {
	prefix string
}

func (l logger) Println(v ...interface{}) {
	log.Println(append([]interface{}{"[" + l.prefix + "] "}, v...)...)
}

func (l logger) Printf(format string, v ...interface{}) {
	log.Printf("["+l.prefix+"] "+format, v...)
}
