package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type Config struct {
	Basic     BasicConfig     `yaml:"basic"`
	DataPlane DataPlaneConfig `yaml:"dataPlane"`
	Resource  ResourceConfig  `yaml:"resources"`
	TestCases []string        `yaml:"testCases"`
}

type BasicConfig struct {
	LocalN4Ip string `yaml:"localN4Ip" validate:"required,ip"`
	UpfN4Ip   string `yaml:"upfN4Ip" validate:"required,ip"`
}

type DataPlaneConfig struct {
	GnbIp string `yaml:"gnbIp" validate:"required,ip"`
	N3Ip  string `yaml:"n3Ip" validate:"required,ip"`
	N6Ip  string `yaml:"n6Ip" validate:"required,ip"`
	DnIp  string `yaml:"dnIp" validate:"required,ip"`
}

type ResourceConfig struct {
	QueueSize uint16 `yaml:"queueSize" validate:"required,min=10,max=65535"`
	StartUeIp string `yaml:"startUeIp" validate:"required,ip"`
	StartSeId uint64 `yaml:"startSeId" validate:"required,min=1"`
	StartTeId uint32 `yaml:"startTeId" validate:"required,min=1"`
}

func (c *Config) LoadConfig(path string) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	if err = yaml.Unmarshal(data, c); err != nil {
		return fmt.Errorf("unmarshal config file failed: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {

	err := validate.Struct(c)
	if err != nil {
		var errMsg string
		for _, err := range err.(validator.ValidationErrors) {
			errMsg += "param: " + err.StructField() + " tag: " + err.Tag() + " validate failed"
		}
		return errors.New(errMsg)
	}

	return nil
}
