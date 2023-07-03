package internal

import (
	"log"
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	MQTT struct {
		URL               string `yaml:"URL"`
		KeepAlive         uint   `yaml:"KeepAlive" default:"60"`
		ConnectTimeout    uint   `yaml:"ConnectTimeout" default:"3"`
		ConnectRetryDelay uint   `yaml:"ConnectRetryDelay" default:"15"`

		ClientId          string `yaml:"ClientId"`
		Username          string `yaml:"Username" `
		Password          string `yaml:"Password" `
		CommandTopic      string `yaml:"CommandTopic"`
		StateTopic        string `yaml:"StateTopic"`
		AvailabilityTopic string `yaml:"AvailabilityTopic"`
	} `yaml:"MQTT"`
	Serial struct {
		Port         string `yaml:"Port"`
		OpenCommand  string `yaml:"OpenCommand"`
		CloseCommand string `yaml:"CloseCommand"`
		StopCommand  string `yaml:"StopCommand"`
		OpenTime     uint   `yaml:"OpenTime"`
		CloseTime    uint   `yaml:"CloseTime"`
	} `yaml:"Serial"`
	API struct {
		Listen string `yaml:"Listen"`
	} `yaml:"API"`
	Debug bool `yaml:"Debug" default:"false"`
}

func LoadConfig(configFile string) *AppConfig {
	appConfig := AppConfig{}
	if err := defaults.Set(&appConfig); err != nil {
		log.Fatalf("[Core] Set default config value failed: %s\n", err)
	}

	fileData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("[Core] Read config file failed: %s\n", err)
	}

	err = yaml.Unmarshal(fileData, &appConfig)
	if err != nil {
		log.Fatalf("[Core] Parse YAML confif failed: %s\n", err)
	}

	verifyConfig(&appConfig)
	return &appConfig
}

func verifyConfig(appConfig *AppConfig) {
	if appConfig.MQTT.URL == "" {
		log.Fatalln("[Core] Missing required argument: MQTT.URL")
	} else if appConfig.MQTT.ClientId == "" {
		log.Fatalln("[Core] Missing required argument: MQTT.ClientId")
	} else if appConfig.MQTT.CommandTopic == "" {
		log.Fatalln("[Core] Missing required argument: MQTT.CommandTopic")
	} else if appConfig.MQTT.StateTopic == "" {
		log.Fatalln("[Core] Missing required argument: MQTT.StateTopic")
	} else if appConfig.MQTT.AvailabilityTopic == "" {
		log.Fatalln("[Core] Missing required argument: MQTT.AvailabilityTopic")
	} else if appConfig.Serial.Port == "" {
		log.Fatalln("[Core] Missing required argument: Serial.Port")
	} else if appConfig.Serial.OpenCommand == "" {
		log.Fatalln("[Core] Missing required argument: Serial.OpenCommand")
	} else if appConfig.Serial.CloseCommand == "" {
		log.Fatalln("[Core] Missing required argument: Serial.CloseCommand")
	} else if appConfig.Serial.StopCommand == "" {
		log.Fatalln("[Core] Missing required argument: Serial.StopCommand")
	}
}
