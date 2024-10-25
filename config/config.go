package config

type ConifgService interface {
	GetConfig(...string) *Config
}

type (
	Config struct {
		Data Data `yaml:"data"`
	}

	Data struct {
		FilePath string `yaml:"filepath"`
	}
)
