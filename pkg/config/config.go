package config

var C *Config = &Config{
	Key:    "",
	Secret: "",
}

type Config struct {
	Key    string
	Secret string
}
