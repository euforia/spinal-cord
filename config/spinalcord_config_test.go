package config

import (
	"encoding/json"
	"testing"
)

func Test_LoadConfigFromTomlFile(t *testing.T) {

	cfg, err := LoadConfigFromTomlFile(testSpinalcordConfig)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	//t.Logf("%#v\n", cfg.Inputs)
	b, _ := json.MarshalIndent(&cfg, "", "  ")
	t.Logf("%s\n", b)
}
