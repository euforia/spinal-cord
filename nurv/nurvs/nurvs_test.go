package nurvs

import (
	"github.com/euforia/spinal-cord/config"
	"github.com/euforia/spinal-cord/logging"
	"os"
	"testing"
)

var testLogger *logging.Logger = logging.NewLogger(os.Stdout, os.Stdout, os.Stdout, os.Stdout, os.Stderr)

var testNurvType string = "amqp"
var testConfigFile string = "/Users/abs/workbench/GoLang/src/github.com/euforia/spinal-cord/nurv-amqp.json"

func Test_LoadNurv(t *testing.T) {
	testConfig, err := config.LoadNurvConfigFromFile(testNurvType, testConfigFile)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	nrv, err := LoadNurv(testConfig, testLogger)
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	t.Logf("%#v", nrv)
}
