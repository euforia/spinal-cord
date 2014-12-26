package decoders

import (
	"testing"
)

func Test_LoadDecoderByName(t *testing.T) {
	resp, err := LoadDecoderByName("OpenStackAMQPDecoder")
	if err != nil {
		t.Errorf("%s", err)
		t.FailNow()
	}
	t.Logf("%v", resp)
}
