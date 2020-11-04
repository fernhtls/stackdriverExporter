package prometheusOutput

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	o := OutputConfig{
		BaseHandlerPath: "/metrics",
		Port: 2100,
	}
	err := o.ValidateConfig()
	assert.NoError(t, err)
}

func TestValidateConfigWrongHandlerPath(t *testing.T) {
	o := OutputConfig{
		BaseHandlerPath: "WRONGPATH",
		Port: 2100,
	}
	err := o.ValidateConfig()
	assert.Error(t, err)
	// Blank HandlerPath
	o = OutputConfig{
		BaseHandlerPath: "",
		Port: 2100,
	}
	err = o.ValidateConfig()
	assert.Error(t, err)
}

func TestValidateConfigPortWithZero(t *testing.T) {
	o := OutputConfig{
		BaseHandlerPath: "/metrics",
		Port: 0,
	}
	err := o.ValidateConfig()
	assert.Error(t, err)
}