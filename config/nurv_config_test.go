package config

import (
	"testing"
)

var testNurvType string = "amqp"
var testConfigFile string = "/Users/abs/workbench/GoLang/src/github.com/euforia/spinal-cord/nurv-amqp.json"
var testSpinalcordConfig string = "/Users/abs/workbench/GoLang/src/github.com/euforia/spinal-cord/spinal-cord.toml"

func Test_LoadNurvConfigFromFile_AMQP(t *testing.T) {
	var (
		err error
	)

	if _, err = LoadNurvConfigFromFile("test", testConfigFile); err == nil {
		t.Errorf("error checking failed: %s", err)
	}

	config, err := LoadNurvConfigFromFile(testNurvType, testConfigFile)
	if err != nil {
		t.Errorf("Failed to load config: %s %s", testConfigFile, err)
	} else {
		otype, valid := config.TypeConfig.(*AMQPNurvConfig)
		if !valid {
			t.Errorf("%#v", otype)
			t.FailNow()
		}

		if otype.RoutingKey != "notifications.info" {
			t.Errorf("routing key mismatch: %s", otype.RoutingKey)
		}

		t.Logf("%#v", config)
	}
}

func Test_NewAMQPConfig(t *testing.T) {
	NewAMQPNurvConfig()
}
