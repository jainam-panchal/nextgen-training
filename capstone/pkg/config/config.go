package config

type Config struct {
	Address string
}

func Default() *Config {
	return &Config{Address: ":8080"}
}
