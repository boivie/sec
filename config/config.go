package config

import (
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"os"
)

var log = logging.MustGetLogger("lovebeat")

type Config struct {
	Http ConfigBind
}

type ConfigBind struct {
	Listen string
}


func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func readFile(conf *Config, fname string) {
	if e, _ := exists(fname); e {
		log.Info("Reading configuration file %s", fname)
		if _, err := toml.DecodeFile(fname, conf); err != nil {
			log.Error("Failed to parse configuration file %s", fname, err)
		}
	}
}

func ReadConfig(fname string) Config {
	var conf = Config{
		Http: ConfigBind{
			Listen: ":8080",
		},
	}
	readFile(&conf, fname)
	return conf
}
