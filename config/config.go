package config

type ConifgService interface {
	GetConfig(...string) *Config
}

type (
	Config struct {
		Data       Data       `yaml:"data"`
		HttpServer HttpServer `yaml:"httpserver"`
	}

	Data struct {
		FilePath string `yaml:"filepath"`
	}

	HttpServer struct {
		Port string `yaml:"port"`
	}
)
