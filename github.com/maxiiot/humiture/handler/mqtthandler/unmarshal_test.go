package mqtthandler

import (
	"encoding/hex"
	"testing"
)

func Test_decodeHumitures(t *testing.T) {
	tdata, err := hex.DecodeString("ff02055801045be3da8002050113ffffa8030330ffa804032fffa8050300ffa8ff00")
	if err != nil {
		t.Error(err)
	}
	hums, err := decodeHumitures(tdata)
	if err != nil {
		t.Error(err)
	}

	t.Log(hums)
}
