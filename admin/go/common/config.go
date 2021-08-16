package common

type DatabaseConfig struct {
	Host        string
	User        string
	Password    string
	Base        string
	Timeout     int
	MaxLifetime int `yaml:"max_lifetime"`
	MaxConn     int `yaml:"max_conn"`
}

type CommonConfig struct {
	Database DatabaseConfig
}

func CommonInitialize(c CommonConfig) {
	DB = OpenDatabase(c.Database)
}
