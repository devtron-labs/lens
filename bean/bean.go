package bean

type GitSensorProtocolConfig struct {
	Protocol string `env:"GIT_SENSOR_PROTOCOL" envDefault:"GRPC"`
}
