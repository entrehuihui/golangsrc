package handler

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"

	"github.com/brocaar/lorawan"
	"github.com/gin-gonic/gin"
	"github.com/maxiiot/humiture/common"
	"github.com/maxiiot/humiture/handler/mqtthandler"
	log "github.com/sirupsen/logrus"
)

type HumitureSet struct {
	DevEUI         string `json:"dev_eui"`
	TemperatureMax int8   `json:"temp_max"`
	TemperatureMin int8   `json:"temp_min"`
	HumidityMax    byte   `json:"hum_max"`
	HumidityMin    byte   `json:"hum_min"`
}

type IntervalSet struct {
	DevEUI   string `json:"dev_eui"`
	Interval byte   `json:"interval"`
}

type DeviceEUISet struct {
	DevEUI    string `json:"dev_eui"`
	DevEUINew string `json:"dev_eui_new"`
	ADDR      string `json:"addr"`
}

type ClassSet struct {
	DevEUI string `json:"dev_eui"`
	Class  string `json:"class"`
}

var regex = regexp.MustCompile("[0-9a-fA-F]{16}")

func DownlinkTime(c *gin.Context) {
	devEUI := c.Param("dev_eui")
	var dev_eui lorawan.EUI64
	if !regex.MatchString(devEUI) {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error", nil)
		c.Abort()
		return
	}
	if err := dev_eui.UnmarshalText([]byte(devEUI)); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error: %s", err)
		c.Abort()
		return
	}
	// 解决A类设备时间差问题，采用异步处理
	if _, ok := common.SyncDevTime[dev_eui]; ok {
		common.SyncDevTime[dev_eui] = true
	}

	// 立即同步服务器时间,A类设备会有时间差.
	// mqtthandler.SyncTime(dev_eui)

	ResponseJSON(c, http.StatusOK, "success", nil)
}

//DownlinkSet hex format:
// header code hum_max temp_max hum_min temp_min tail
//  1       1    1        1       1        1     1
func DownlinkCriticalSet(c *gin.Context) {
	var hs HumitureSet
	var dev_eui lorawan.EUI64
	if err := c.BindJSON(&hs); err != nil {
		ResponseJSON(c, http.StatusBadRequest, err.Error(), nil)
		c.Abort()
		return
	}

	if hs.TemperatureMax <= hs.TemperatureMin {
		ResponseJSON(c, http.StatusBadRequest, "max temperature must greater than min temperature.", nil)
		c.Abort()
		return
	}

	if hs.HumidityMax <= hs.HumidityMin {
		ResponseJSON(c, http.StatusBadRequest, "max humidity must greater than min humidity.", nil)
		c.Abort()
		return
	}

	if !regex.MatchString(hs.DevEUI) {
		ResponseJSON(c, http.StatusBadRequest, "device eui is empty or format error", nil)
		c.Abort()
		return
	}

	if err := dev_eui.UnmarshalText([]byte(hs.DevEUI)); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error: %s", err)
		c.Abort()
		return
	}
	buf := bytes.NewBuffer([]byte{})
	buf.Write([]byte{0xff, 0x11})

	buf.WriteByte(hs.HumidityMax)
	buf.WriteByte(byte(hs.TemperatureMax))
	buf.WriteByte(hs.HumidityMin)
	buf.WriteByte(byte(hs.TemperatureMin))
	buf.WriteByte(0xff)
	log.WithField("devEUI", hs.DevEUI).Debugf("downlink:%x", buf.Bytes())
	// bs64_data := base64.StdEncoding.EncodeToString(buf.Bytes())
	mqtthandler.PubChan <- mqtthandler.PublishChan{
		DevEUI: dev_eui,
		Payload: mqtthandler.DataDownPayload{
			DevEUI:    dev_eui,
			Confirmed: false,
			FPort:     10,
			Data:      buf.Bytes(),
		},
	}
	ResponseJSON(c, http.StatusOK, "success", nil)
}

// hex format:
// header code interval tail
//   1      1     1      1
func DownlinkIntervalSet(c *gin.Context) {
	var set IntervalSet
	var dev_eui lorawan.EUI64
	err := c.BindJSON(&set)
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, err.Error(), nil)
		c.Abort()
		return
	}
	if !regex.MatchString(set.DevEUI) {
		ResponseJSON(c, http.StatusBadRequest, "device eui is empty or format error", nil)
		c.Abort()
		return
	}
	if err := dev_eui.UnmarshalText([]byte(set.DevEUI)); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error: %s", err)
		c.Abort()
		return
	}

	b := []byte{0xff, 0x12, set.Interval, 0xff}
	log.WithField("devEUI", set.DevEUI).Debugf("downlink:%x", b)
	// bs64_data := base64.StdEncoding.EncodeToString(b)
	mqtthandler.PubChan <- mqtthandler.PublishChan{
		DevEUI: dev_eui,
		Payload: mqtthandler.DataDownPayload{
			DevEUI:    dev_eui,
			Confirmed: false,
			FPort:     10,
			Data:      b,
		},
	}
	ResponseJSON(c, http.StatusOK, "success", nil)
	return
}

// hex format:
// header  code dev_eui dev_addr tail
//   1       1     8       4       1
func DownlinkEUISet(c *gin.Context) {
	var set DeviceEUISet
	var dev_eui lorawan.EUI64
	err := c.BindJSON(&set)
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, err.Error(), nil)
		c.Abort()
		return
	}

	if !regex.MatchString(set.DevEUI) {
		ResponseJSON(c, http.StatusBadRequest, "device eui is empty or format error", nil)
		c.Abort()
		return
	}

	if err := dev_eui.UnmarshalText([]byte(set.DevEUI)); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error: %s", err)
		c.Abort()
		return
	}
	var devEUI lorawan.EUI64
	var addr lorawan.DevAddr
	err = devEUI.UnmarshalText([]byte(set.DevEUINew))
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, err.Error(), nil)
		c.Abort()
		return
	}
	err = addr.UnmarshalText([]byte(set.ADDR))

	buf := bytes.NewBuffer([]byte{0xff, 0x13})
	buf.Write(devEUI[:])
	buf.Write(addr[:])
	buf.WriteByte(0xff)
	log.WithField("devEUI", set.DevEUI).Debugf("downlink:%x", buf.Bytes())
	//bs64_data := base64.StdEncoding.EncodeToString(buf.Bytes())
	mqtthandler.PubChan <- mqtthandler.PublishChan{
		DevEUI: dev_eui,
		Payload: mqtthandler.DataDownPayload{
			DevEUI:    dev_eui,
			Confirmed: false,
			FPort:     10,
			Data:      buf.Bytes(),
		},
	}
	ResponseJSON(c, http.StatusOK, "success", nil)
	return
}

//
func DownlinkClassSet(c *gin.Context) {
	var set ClassSet
	var dev_eui lorawan.EUI64
	err := c.BindJSON(&set)
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, err.Error(), nil)
		c.Abort()
		return
	}

	if !regex.MatchString(set.DevEUI) {
		ResponseJSON(c, http.StatusBadRequest, "device eui is empty or format error", nil)
		c.Abort()
		return
	}

	if err := dev_eui.UnmarshalText([]byte(set.DevEUI)); err != nil {
		ResponseJSON(c, http.StatusBadRequest, "device eui format error: %s", err)
		c.Abort()
		return
	}
	var class byte
	if strings.ToUpper(set.Class) == "CLASSA" {
		class = 0
	} else if strings.ToUpper(set.Class) == "CLASSC" {
		class = 1
	} else {
		ResponseJSON(c, http.StatusBadRequest, "invalid class value.", nil)
		c.Abort()
		return
	}
	b := []byte{0xff, 0x14, class, 0xff}
	//bs64_data := base64.StdEncoding.EncodeToString(b)
	mqtthandler.PubChan <- mqtthandler.PublishChan{
		DevEUI: dev_eui,
		Payload: mqtthandler.DataDownPayload{
			DevEUI:    dev_eui,
			Confirmed: false,
			FPort:     0,
			Data:      b,
		},
	}
	ResponseJSON(c, http.StatusOK, "success", nil)
	return
}
