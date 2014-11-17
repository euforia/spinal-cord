package libs

import(
    "io/ioutil"
    "encoding/json"
)

type Config struct {
    Namespace string       `json:"namespace"`
    SpinalCord string      `json:"spinal_cord"`
    LogLevel string        `json:"log_level"`
    NurvType string        `json:"nurv_type"`
    TypeConfig interface{} `json:"nurv_config"`
    Version string         `json:"version"`
}

func NewConfig(ns, spinalcord, loglevel, nurvtype, version string) *Config {
    return &Config{ns, spinalcord, loglevel, nurvtype, nil, version}
}

func LoadConfigFromFile(filepath string, cfg *Config) error {

    contents, err := ioutil.ReadFile(filepath)
    if err != nil {
        return err
    }
    return json.Unmarshal(contents, &cfg)
}

type AMQPConfig struct {
    URI string         `json:"uri"`
    QueueName string   `json:"queue_name"`
    ExchangeType string `json:"exchange_type"`
    RoutingKey string  `json:"routing_key"`
    ConsumerTag string `json:"consumer_tag"`
    Exchanges []string `json:"exchanges"`
}
func NewAMPQConfig() *AMQPConfig {
    return &AMQPConfig{"","","","","", make([]string,0)}
}
