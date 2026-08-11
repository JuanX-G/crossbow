package crossbow

import (
	"testing"
)

func TestDefaultConfigWorkers(t *testing.T) {
	tests := map[int]uint{ // map input to expected value of Workers in the output struct
		0:  1,
		6:  6,
		-1: 1, // sentinel for in passing nothing
	}
	for in, ex := range tests {
		if in == -1 { // check default to 1 when to argument is provided
			cfg := MakeDefaultServerConfig()
			if cfg.Workers != 1 {
				t.Fatalf("passed no workers argument to MakeDefaultServerConfig, expected Workers: %d in the final config, found: %d", ex, cfg.Workers)
			}
			continue
		}
		cfg := MakeDefaultServerConfig(uint(in))
		if cfg.Workers != ex {
			t.Fatalf("passed %d as workers argument to MakeDefaultServerConfig, found: %d, in the returned config, expected: %d", in, cfg.Workers, ex)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	tests := map[ServerConfig]ServerConfigError{
		ServerConfig{Workers: 0}:                 ServerConfigError{ErrType: ConfigErrorWorkersZero},
		ServerConfig{MailboxSize: 0, Workers: 1}: ServerConfigError{ErrType: ConfigErrorMailboxSizeZero},
	}
	for cfg, exErr := range tests {
		err := cfg.Validate()
		err2, ok := err.(ServerConfigError)
		if !ok {
			t.Fatalf("Validate() returned error of type: %T, expected: %T", err, ServerConfigError{})
		}
		if err2 != exErr {
			t.Fatalf("Validate() returned error with ErrType: %s, expected ErrType was: %s", err2.ErrType.String(), exErr.ErrType.String())
		}

	}
}
