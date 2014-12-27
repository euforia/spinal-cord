package config

type ReactorConfig struct {
	Enabled       bool          `toml:"enabled" json:"enabled"`
	Port          int           `toml:"port" json:"port"`
	CreateSamples bool          `toml:"create_samples" json:"create_samples"`
	Synapse       SynapseConfig `json:"synapse"`
}

type WebsocketConfig struct {
	Enabled  bool          `toml:"enabled" json:"enabled"`
	Endpoint string        `toml:"endpoint" json:"endpoint"`
	Synapse  SynapseConfig `json:"synapse"`
}
