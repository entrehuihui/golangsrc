package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maxiiot/humiture/myinfluxdb"
	"github.com/maxiiot/humiture/setting"
)

func Index() http.Handler {
	return http.FileServer(http.Dir(setting.Cfg.General.WorkPath + "/ui/"))
}

func GetDeviceChart(c *gin.Context) {
	devEui := c.Param("dev_eui")
	if devEui == "" {
		ResponseJSON(c, http.StatusBadRequest, "device eui is empty", nil)
		c.Abort()
		return
	}
	// logs, err := models.GetHumitures(common.DB, setting.Cfg.General.ChartNum, 0, devEui)
	logs, err := myinfluxdb.GetDeviceChart(setting.Cfg.General.ChartNum, 0, devEui)
	if err != nil {
		ResponseJSON(c, http.StatusInternalServerError, fmt.Sprintf("%s", err), nil)
		c.Abort()
		return
	}
	length := len(logs)
	type Humiture struct {
		Temp        []interface{} `json:"temp"`
		Humidity    []interface{} `json:"humidity"`
		Electricity []interface{} `json:"electricity"`
		UpDate      []string      `json:"up_date"`
	}
	humiture := Humiture{
		Temp:        make([]interface{}, length),
		Humidity:    make([]interface{}, length),
		UpDate:      make([]string, length),
		Electricity: make([]interface{}, length),
	}
	for idx, log := range logs {
		humiture.Temp[idx] = log.Temperature
		humiture.Humidity[idx] = log.Humidity
		humiture.UpDate[idx] = log.UpDate.UTC().Add(time.Hour * 8).Format("2006-01-02 15:04:05")
		humiture.Electricity[idx] = log.Electricity
	}
	ResponseJSON(c, http.StatusOK, "success", humiture)
}

func GetDevices(c *gin.Context) {
	// devs, err := models.GetDevices(common.DB)
	devs, err := myinfluxdb.GetDevices()
	if err != nil {
		ResponseJSON(c, http.StatusInternalServerError, fmt.Sprintf("%s", err), nil)
		c.Abort()
		return
	}
	ResponseJSON(c, http.StatusOK, "success", gin.H{
		"devices": devs,
	})
}

func GetDeviceHistory(c *gin.Context) {
	devEUI := c.Param("dev_eui")
	if devEUI == "" {
		ResponseJSON(c, http.StatusBadRequest, "devices eui empty", nil)
		c.Abort()
		return
	}
	now := time.Now()
	df_start := now.Format(time.RFC3339)
	df_end := now.Format(time.RFC3339)
	start := c.DefaultQuery("start", df_start)
	end := c.DefaultQuery("end", df_end)
	perPage := c.DefaultQuery("per_page", "100")
	page := c.DefaultQuery("page", "1")
	limit, err := strconv.Atoi(perPage)
	if err != nil || limit <= 0 {
		ResponseJSON(c, http.StatusBadRequest, "per_page must greater than 0", nil)
		c.Abort()
		return
	}
	offset, err := strconv.Atoi(page)
	if err != nil || offset <= 0 {
		ResponseJSON(c, http.StatusBadRequest, "page must greater than 0.", nil)
		c.Abort()
		return
	}
	offset = (offset - 1) * limit

	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.UTC
	}
	start_time, err := time.ParseInLocation("2006-01-02 15:04:05", start, location)
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, "start time format error.valid format('2006-01-02 15:04:05')", nil)
		c.Abort()
		return
	}
	end_time, err := time.ParseInLocation("2006-01-02 15:04:05", end, location)
	if err != nil {
		ResponseJSON(c, http.StatusBadRequest, "end time format error. valid format('2006-01-02 15:04:05')", nil)
		c.Abort()
		return
	}

	count := myinfluxdb.GetDeviceHistoryCount(devEUI, start_time, end_time)
	if count != nil {
		ResponseJSON(c, http.StatusInternalServerError, "get humiture history count error", nil)
		c.Abort()
		return
	}
	his, err := myinfluxdb.GetDeviceHistory(limit, offset, devEUI, start_time, end_time)
	if err != nil {
		ResponseJSON(c, http.StatusInternalServerError, "get humiture history error", nil)
		c.Abort()
		return
	}
	// his, err := models.GetHumituresHistory(common.DB, limit, offset, devEUI, start_time, end_time)
	// count, err := models.GetHumituresHistoryCount(common.DB, devEUI, start_time, end_time)
	// if err != nil {
	// 	ResponseJSON(c, http.StatusInternalServerError, "get humiture history count error", nil)
	// 	c.Abort()
	// 	return
	// }
	ResponseJSON(c, http.StatusOK, "success", gin.H{
		"total_count": count,
		"humiture":    his,
	})
}
