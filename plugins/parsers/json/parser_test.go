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
		"conn": "0",
		"status": "1",
		"health": "11"
	}, {
		"name": "qagslb.glb.cisco.com",
		"establishedconn": "0",
		"status": "2",
		"health": "22"
	}]
}
`

func TestGslbData(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "httpjson_gslb_dnsrecords",
		BasePath:   "gslbvserver",
		TagKeys:    []string{"name"},
		FieldMap: map[string]string{
			"status": "float",
			"health": "float",
			"conn":   "",
		},
	}
	metrics, err := parser.Parse([]byte(validGslbData))
	assert.NoError(t, err)
	assert.Len(t, metrics, 2)

	assert.Equal(t, "httpjson_gslb_dnsrecords", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"status": float64(1),
		"health": float64(11),
		"conn":   "0",
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{"name": "capi.glb.cisco.com"}, metrics[0].Tags())

	assert.Equal(t, "httpjson_gslb_dnsrecords", metrics[1].Name())
	assert.Equal(t, map[string]string{"name": "qagslb.glb.cisco.com"}, metrics[1].Tags())

}
