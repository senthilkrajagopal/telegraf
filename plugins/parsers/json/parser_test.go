package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const validGslbData = `
{
	"errorcode": 0,
	"message": "Done",
	"severity": "NONE",
	"gslbvserver": [{
		"name": "capi.glb.cisco.com",
		"conn": 0,
		"status": "1",
		"health": "0"
	}, {
		"name": "qagslb.glb.cisco.com",
		"establishedconn": 0,
		"status": "1",
		"health": "0"
	}]
}
`

func TestGslbData(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "httpjson_gslb_dnsrecords",
		BasePath:   "gslbvserver",
		TagKeys:    []string{"name"},
	}
	metrics, err := parser.Parse([]byte(validGslbData))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

}
