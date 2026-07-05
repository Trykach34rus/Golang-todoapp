package core_http_server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr           string `envcofing:"ADDR" required:"true"`
	ShutdownTimeout time.Duration `envcofing:"SHUTDOWN_TIMEOUT" default:"30s"`
}


func NewConfig() (Config,error)  {
	var config Config

	if err := envconfig.Process("HTTP",&config); err != nil {
		 return Config{}, fmt.Errorf("process envcofing: %w",err)
	}

	return config,nil

}

func NewConfigMust() Config  {
	config,err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get HTTP config: %w ",err)
		panic(err)
	}

	return config
}



