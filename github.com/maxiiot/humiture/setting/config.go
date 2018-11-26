package setting

import (
	"io/ioutil"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var Cfg *Config
var once sync.Once

type General struct {
	Port        int    `toml:"port"`
	WorkPath    string `toml:"workPath"`
	ChartNum    int    `toml:"chartNum"`
	DSN         string `toml:"dsn"`
	AutoMigrate bool   `toml:"automigrate"`
	LogLevel    int    `toml:"logLevel"`
}

type MqttServer struct {
	Host          string   `toml:"host"`
	UserName      string   `toml:"username"`
	Password      string   `toml:"password"`
	CaFile        string   `toml:"cafile"`
	ApplicationID int      `toml:"applicationID"`
	DevEUI        []string `toml:"devEUI"`
	RxTopic       string   `toml:"rxTopic"`
}

type Influxdb struct {
	Addr      string `toml:"addr"`
	UserName  string `toml:"username"`
	Password  string `toml:"password"`
	Database  string `toml:"database"`
	Precision string `toml:"precision"`
	TableName string `toml:"table"`
}

type Config struct {
	General    `toml:"general"`
	MqttServer `toml:"mqttserver"`
	Influxdb   `toml:"influxdb"`
}

func LoadConfig(paths ...string) error {
	for _, path := range paths {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			log.Warn("read config file error:", err)
			continue
		}
		Cfg = &Config{}
		err = toml.Unmarshal(b, Cfg)
		if err != nil {
			log.Warn("unmarshal config file error:", err)
			continue
		}
		return nil
	}

	return errors.New("load config fatal")
}
