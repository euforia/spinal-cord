package config

type SynapseConfig struct {
	Type       string      `json:"type"`
	URI        string      `json:"uri"`
	TypeConfig interface{} `toml:"config" json:"config"`
}

type ZmqSynapseConfig struct {
	Type string `json:"type"`
}

type RecvSynapseConfig struct {
	Type          string   `toml:"type" json:"type"`
	Subscriptions []string `toml:"subscriptions" json:"subscriptions"`
}
