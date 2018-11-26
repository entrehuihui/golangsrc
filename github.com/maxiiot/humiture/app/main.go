package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/maxiiot/humiture/handler/mqtthandler"
	"github.com/maxiiot/humiture/myinfluxdb"
	"github.com/maxiiot/humiture/routers"
	"github.com/maxiiot/humiture/setting"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func main() {
	log.Fatal(run())
}

func run() error {

	h, err := mqtthandler.NewHandler(setting.Cfg.MqttServer.Host,
		setting.Cfg.MqttServer.UserName, setting.Cfg.MqttServer.Password, setting.Cfg.MqttServer.CaFile,
		setting.Cfg.MqttServer.ApplicationID)
	if err != nil {
		log.Error(err)
	}
	mqtthandler.PubChan = make(chan mqtthandler.PublishChan, 100)
	WS := make(map[*websocket.Conn]struct{})
	go h.HandleRXPackets(WS)
	go h.PublishData()
	defer func() {
		h.Close()
		close(mqtthandler.PubChan)
	}()

	log.Info("start web server:")
	r := gin.New()
	r.Use(gin.Recovery())
	routers.Router(r, WS)
	port := fmt.Sprintf(":%d", setting.Cfg.General.Port)
	log.Info("Now Listening ", port)
	return r.Run(port)
}

func init() {
	err := setting.LoadConfig("../config/app.toml", "/etc/humiture/app.toml")
	if err != nil {
		log.Fatal(err)
	}
	myinfluxdb.MyinfluxdbInit()
	// common.DB, err = dbhelper.OpenDatabase(setting.Cfg.General.DSN)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// common.SyncDevTime = make(map[lorawan.EUI64]bool)
	// err = dbhelper.MigrateHumiture(common.DB)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// devs, err := myinfluxdb.GetDevices()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, dev := range devs {
	// 	var dev_eui lorawan.EUI64
	// 	if err := dev_eui.UnmarshalText([]byte(dev.DevEUI)); err == nil {
	// 		common.SyncDevTime[dev_eui] = false
	// 	}
	// }
	log.SetLevel(log.Level(setting.Cfg.General.LogLevel))
	gin.SetMode(gin.ReleaseMode)
}
