package config

type ResolvePluginsConfig struct {
	Enable  bool
	Plugins map[string]map[string]string
}

type AdminConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
