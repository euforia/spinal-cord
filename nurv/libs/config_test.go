package libs

import(
    "testing"
)

var testConfigFile string = "amqp-nurv_test.conf"

func Test_LoadConfigFromFile_AMQP(t *testing.T) {

    config := NewConfig("","","","","")
    config.TypeConfig = NewAMQPConfig()

    err := LoadConfigFromFile(testConfigFile, config)
    if err != nil {
        t.Errorf("Failed to load config: %s %s", testConfigFile, err)
    } else {
        if otype, valid := config.TypeConfig.(*AMQPConfig); !valid {
            t.Errorf("%#v", otype)
        } else {
            t.Logf("%#v", otype)
        }
    }
}
