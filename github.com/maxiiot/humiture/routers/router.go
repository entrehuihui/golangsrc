package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/maxiiot/humiture/handler"
	"github.com/maxiiot/humiture/setting"
	"golang.org/x/net/websocket"
)

func Router(router *gin.Engine, ws map[*websocket.Conn]struct{}) {
	router.Use(handler.Cors())

	api := router.Group("/api")
	{
		api.GET("/dev/list", handler.GetDevices)
		api.GET("/dev/humiture/:dev_eui", handler.GetDeviceChart)
		api.GET("/dev/history/:dev_eui", handler.GetDeviceHistory)
		api.GET("/dev/downlink/:dev_eui", handler.DownlinkTime)
		api.POST("/dev/downlink/set", handler.DownlinkCriticalSet)
		api.POST("/dev/downlink/interval", handler.DownlinkIntervalSet)
		api.POST("/dev/downlink/deveui",handler.DownlinkEUISet)
		api.POST("/dev/downlink/class",handler.DownlinkClassSet)
	}
	router.Static("/ui", setting.Cfg.General.WorkPath)
	router.StaticFile("/", setting.Cfg.General.WorkPath+"/index.html")
	router.Handle("GET", "/alarm", handler.AlarmHandler(ws))
}
