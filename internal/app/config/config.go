package config

type Config struct {
	serverURL     string
	baseReturnURL string
}

func (c *Config) SetServerURL(flagValue string) error {
	c.serverURL = flagValue
	return nil
}

func (c *Config) SetBaseReturnURL(flagValue string) error {
	c.baseReturnURL = flagValue
	return nil
}

func (c *Config) GetServerURL() string {
	return c.serverURL
}

func (c *Config) GetBaseReturnURL() string {
	return c.baseReturnURL
}

func GetDefault() *Config {
	config := Config{
		serverURL:     "localhost:8080",
		baseReturnURL: "http://localhost:8080",
	}
	return &config
}
