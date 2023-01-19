package config

type ResolveModulesConfig struct {
	Enable  bool                         `yaml:"enable"`
	Modules map[string]map[string]string `yaml:"modules"`
}

type AdminConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}
