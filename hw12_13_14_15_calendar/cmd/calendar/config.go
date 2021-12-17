package main

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Host        string
	Port        int
	Logger      LoggerConf
	StorageType string
	Database    DatabaseConf
}

type LoggerConf struct {
	Level string
}

type DatabaseConf struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

func NewConfig() Config {
	return Config{}
}
