package setting

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_LoadConfig(t *testing.T) {
	Convey("test load config", t, func() {
		LoadConfig("../config/app.toml")
		So(Cfg.General.Port, ShouldEqual, 9000)
		So(Cfg.MqttServer.ApplicationID, ShouldEqual, 0)
		So(Cfg.MqttServer.DevEUI[0], ShouldEqual, "2018041332000001")
		So(Cfg.General.WorkPath, ShouldEqual, "../")
		So(Cfg.General.ChartNum, ShouldEqual, 30)
		So(Cfg.General.AutoMigrate, ShouldEqual, true)
	})
}
