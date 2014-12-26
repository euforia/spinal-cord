package nurvs

import (
	"github.com/euforia/spinal-cord/config"
	"testing"
	"time"
)

func Test_NewAMQPNurv(t *testing.T) {

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

	if err = nrv.Start(); err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}

	time.Sleep(1)

	if err = nrv.Stop(); err != nil {
		t.Errorf("%s", err)
	}
}
