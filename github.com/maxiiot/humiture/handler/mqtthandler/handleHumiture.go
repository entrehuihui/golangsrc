package mqtthandler

import (
	"fmt"
	"time"

	"github.com/maxiiot/humiture/myinfluxdb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func handleHumitures(rxPacket *DataUpPayloadChan, ws map[*websocket.Conn]struct{}) {
	if len(rxPacket.Data) < 12 {
		log.Errorf("Humiture data length expect:%d", len(rxPacket.Data))
		return
	}
	// 单组温湿度数据处理
	if rxPacket.Data[0] == 0xff && rxPacket.Data[1] == 0x01 {
		h, err := decodeHumiture(rxPacket.Data)
		if err != nil {
			log.Error(err)
			return
		}
		err = handleHumiture(rxPacket.DevEUI, rxPacket.DevName, ws, h)
		if err != nil {
			log.Error(err)
			return
		}
	} else if rxPacket.Data[0] == 0xff && rxPacket.Data[1] == 0x02 { //多组温湿度数据处理
		hums, err := decodeHumitures(rxPacket.Data)
		if err != nil {
			log.WithError(err).Error("decodeHumitures error")
			return
		}
		for _, hum := range hums {
			err := handleHumiture(rxPacket.DevEUI, rxPacket.DevName, ws, hum)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}
}

func handleHumiture(devEUI, devName string, ws map[*websocket.Conn]struct{}, h humiture) error {
	log.WithFields(log.Fields{
		"devEUI":      devEUI,
		"temperature": h.temp,
		"humidity":    h.hum,
		"electricity": h.elec,
		"up_date":     h.up_date.Format("2006-01-02 15:04:05"),
	}).Debug("decode humiture result")
	// fmt.Println("保存数据进post数据库")
	// item := &models.HumitureLog{
	// 	DevEUI:      devEUI,
	// 	DevName:     devName,
	// 	Temperature: h.temp,
	// 	Humidity:    h.hum,
	// 	Electricity: h.elec,
	// 	UpDate:      h.up_date,
	// 	CreatedAt:   time.Now(),
	// }
	// err := models.InsertHumiture(common.DB, item)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("保存数据进influxdb数据库")
	//保存数据进influxdb数据库
	itemMaps := []myinfluxdb.SaveDataInfo{
		myinfluxdb.SaveDataInfo{
			Tags: map[string]string{
				"DevEUI":  devEUI,
				"DevName": devName,
			},
			Data: map[string]interface{}{
				"Temperature": h.temp,
				"Humidity":    h.hum,
				"Electricity": h.elec,
				"UpDate":      time.Now().Unix(),
			},
			Time: h.up_date,
		},
	}
	myinfluxdb.SaveData(itemMaps)
	if h.alarm > 0 {
		/// websocket write alarm info.
		info := unmarshalAlarm(h.alarm)
		for ws_conn, _ := range ws {
			_, err := ws_conn.Write([]byte(fmt.Sprintf("%s=> %s: %s", h.up_date.UTC().Add(time.Hour*8).Format("2006-01-02 15:04:05"), devName, info)))
			if err != nil {
				delete(ws, ws_conn)
				return errors.Errorf("write websocket client %s,error: %s", ws_conn.RemoteAddr(), err)
			}
		}
	}
	return nil
}
