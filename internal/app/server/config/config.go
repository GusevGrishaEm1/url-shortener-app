package config

type Config struct {
	serverURL     string
	baseReturnURL string
}

func (c *Config) SetServerURL(flagValue string) error {
	c.serverURL = flagValue
	return nil
}

func (c *Config) SetBaseReturnURLURL(flagValue string) error {
	c.baseReturnURL = flagValue
	return nil
}

func (c *Config) GetServerURL() string {
	return c.serverURL
}

func (c *Config) GetBaseReturnURL() string {
	return c.baseReturnURL
}
