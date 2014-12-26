package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type NurvConfig struct {
	Namespace  string        `json:"namespace"`
	SpinalCord string        `json:"spinal_cord"`
	LogLevel   string        `json:"log_level"`
	Type       string        `json:"type"`
	Decoder    string        `json:"decoder"`
	TypeConfig interface{}   `json:"type_config"`
	Synapse    SynapseConfig `json:"synapse"`
	Version    string        `json:"version"`
}

func LoadNurvConfigFromFile(nurvType string, filepath string) (*NurvConfig, error) {
	ncfg := NurvConfig{}

	switch nurvType {
	case "amqp":
		ncfg.TypeConfig = NewAMQPNurvConfig()
		break
	default:
		return &ncfg, fmt.Errorf("Invalid nurv type: %s", nurvType)
		break
	}

	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return &ncfg, err
	}
	if err := json.Unmarshal(contents, &ncfg); err != nil {
		return &ncfg, err
	}

	return &ncfg, nil
}

type AMQPNurvConfig struct {
	URI          string   `json:"uri"`
	QueueName    string   `json:"queue_name"`
	ExchangeType string   `json:"exchange_type"`
	RoutingKey   string   `json:"routing_key"`
	ConsumerTag  string   `json:"consumer_tag"`
	Exchanges    []string `json:"exchanges"`
}

func NewAMQPNurvConfig() *AMQPNurvConfig {
	return &AMQPNurvConfig{"", "", "", "", "", make([]string, 0)}
}
