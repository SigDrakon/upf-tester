package config

import (
	"strings"
	"testing"
)

func TestConfig_LoadConfig_FileNotFound(t *testing.T) {

	var cfg Config
	err := cfg.LoadConfig("non-existent-config.yaml")
	if err == nil {
		t.Errorf("LoadConfig should fail")
		return
	}

	if !strings.Contains(err.Error(), "open non-existent-config.yaml: no such file or directory") {
		t.Errorf("LoadConfig should fail")
	}
}

func TestConfig_LoadConfig_Validate(t *testing.T) {

	tests := []struct {
		name              string
		yamlPath          string
		expectLoadErr     bool
		expectValidateErr bool
		errStr            string
	}{
		{
			name:              "loss basic.localN4Ip",
			yamlPath:          "./testdata/loss_basic.localN4Ip.yaml",
			expectLoadErr:     false,
			expectValidateErr: true,
			errStr:            "param: LocalN4Ip tag: required validate failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			loadErr := cfg.LoadConfig(tt.yamlPath)
			if (loadErr != nil) != tt.expectLoadErr {
				t.Errorf("LoadConfig() error = %v, expectLoadErr %v", loadErr, tt.expectLoadErr)
				return
			}

			validateErr := cfg.Validate()
			if (validateErr != nil) != tt.expectValidateErr {
				t.Errorf("Validate() error = %v, expectValidateErr %v", validateErr, tt.expectValidateErr)
				return
			}

			if tt.expectValidateErr {
				if !strings.EqualFold(validateErr.Error(), tt.errStr) {
					t.Errorf("Validate() error = %v, expectValidateErr %v", validateErr, tt.errStr)
					return
				}
			}
		})
	}
}
