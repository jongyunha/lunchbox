package config

import (
	"os"
	"time"

	"github.com/jongyunha/lunchbox/internal/rpc"
	"github.com/jongyunha/lunchbox/internal/web"
	"github.com/kelseyhightower/envconfig"
	"github.com/stackus/dotenv"
)

type (
	PGConfig struct {
		User              string `required:"true" envconfig:"PG_USER"`
		Password          string `required:"true" envconfig:"PG_PASSWORD"`
		Host              string `required:"true" envconfig:"PG_HOST"`
		Port              string `required:"true" envconfig:"PG_PORT"`
		DBName            string `required:"true" envconfig:"PG_DBNAME"`
		SSLMode           string `required:"true" envconfig:"PG_SSLMODE"`
		MaxConns          int32  `default:"25" envconfig:"PG_MAX_CONNS"`
		MinConns          int32  `default:"5" envconfig:"PG_MIN_CONNS"`
		MaxConnLifetime   int    `default:"3600" envconfig:"PG_MAX_CONN_LIFETIME"`
		MaxConnIdleTime   int    `default:"1800" envconfig:"PG_MAX_CONN_IDLE_TIME"`
		HealthCheckPeriod int    `default:"60" envconfig:"PG_HEALTH_CHECK_PERIOD"`
	}

	AppConfig struct {
		Environment     string
		LogLevel        string `envconfig:"LOG_LEVEL" default:"DEBUG"`
		PG              PGConfig
		Web             web.WebConfig
		Rpc             rpc.RpcConfig
		ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	}
)

func InitConfig() (cfg AppConfig, err error) {
	if err = dotenv.Load(dotenv.EnvironmentFiles(os.Getenv("ENVIRONMENT"))); err != nil {
		return
	}

	err = envconfig.Process("", &cfg)
	return
}
