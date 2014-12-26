package config

type SynapseConfig struct {
	Type       string      `json:"type"`
	SpinalUri  string      `json:"spinal_uri"`
	TypeConfig interface{} `json:"config"`
}

type ZmqSynapseConfig struct {
	Type string `json:"type"`
}
