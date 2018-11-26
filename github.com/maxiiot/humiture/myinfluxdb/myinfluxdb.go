package myinfluxdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/maxiiot/humiture/setting"
)

// Conn ..
var Conn influxdb.Client

//Bp ..
var Bp influxdb.BatchPoints
var bpNUm = 0

//MyinfluxdbInit ..
func MyinfluxdbInit() {
	conn, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     setting.Cfg.Influxdb.Addr,
		Username: setting.Cfg.Influxdb.UserName,
		Password: setting.Cfg.Influxdb.Password,
	})
	if err != nil {
		log.Fatal(err)
	}
	Conn = conn
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  setting.Cfg.Influxdb.Database,
		Precision: setting.Cfg.Influxdb.Precision,
	})
	if err != nil {
		log.Fatal(err)
	}
	Bp = bp
	go writeBp()
}

func writeBp() {
	for {
		if bpNUm > 0 {
			if err := Conn.Write(Bp); err != nil {
				fmt.Println("inluxdb error : ", err)
			}
			bpNUm = 0
		}
		time.Sleep(time.Second * 60)
		fmt.Println("influxdb will check write")
	}
}

//SaveDataInfo ..
type SaveDataInfo struct {
	Tags map[string]string
	Data map[string]interface{}
	Time time.Time
}

//SaveData ..
func SaveData(saveDataInfos []SaveDataInfo) error {
	for _, saveDataInfo := range saveDataInfos {
		pt, err := influxdb.NewPoint(setting.Cfg.Influxdb.TableName, saveDataInfo.Tags, saveDataInfo.Data, saveDataInfo.Time)
		if err != nil {
			fmt.Println("inluxdb error : ", err)
			return err
		}
		Bp.AddPoint(pt)
	}
	// if err := Conn.Write(Bp); err != nil {
	// 	fmt.Println("inluxdb error : ", err)
	// 	return err
	// }
	bpNUm++
	return nil
}

//HumitureLog ..
type HumitureLog struct {
	ID          int64       `db:"id" json:"id"`
	DevEUI      string      `db:"dev_eui" json:"dev_eui"`
	DevName     string      `db:"dev_name" json:"dev_name"`
	Temperature interface{} `db:"temperature" json:"temperature"`
	Humidity    interface{} `db:"humidity" json:"humidity"`
	Electricity interface{} `db:"electricity" json:"electricity"`
	UpDate      time.Time   `db:"up_date" json:"up_date"`
	CreatedAt   time.Time   `db:"created_at"`
}

//HumitureDevice ..
type HumitureDevice struct {
	DevEUI  string `db:"dev_eui" json:"dev_eui"`
	DevName string `db:"dev_name" json:"dev_name"`
}

//GetHumitures ..
// func GetHumitures(limit, offset int, devEUI string) {
// 	q := influxdb.Query{
// 		Command:  "select * from test order by time desc limit 10",
// 		Database: "entre",
// 	}
// 	response, err := Conn.Query(q)
// 	if err == nil {
// 		if response.Error() != nil {
// 			return
// 		}
// 		fmt.Println(*response)
// 	} else {
// 		return
// 	}
// 	for i, row := range response.Results[0].Series[0].Values {
// 		t, err := time.Parse(time.RFC3339, row[0].(string))
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		t = t.UTC().Add(time.Hour * 8)
// 		valu := row[1]
// 		fmt.Println(i, t, valu, row[2], row[3], row[4])
// 	}
// }

//GetDevices ..
func GetDevices() ([]HumitureDevice, error) {
	q := influxdb.Query{
		Command:  fmt.Sprintf("select DevEUI, DevName, UpDate from %s group by DevEUI, DevName limit 1", setting.Cfg.Influxdb.TableName),
		Database: setting.Cfg.Influxdb.Database,
	}
	response, err := Conn.Query(q)
	if err == nil {
		if response.Error() != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	humitureDevices := make([]HumitureDevice, 0)
	for _, Serie := range response.Results[0].Series {
		humitureDevice := HumitureDevice{
			DevEUI:  Serie.Tags["DevEUI"],
			DevName: Serie.Tags["DevName"],
		}
		humitureDevices = append(humitureDevices, humitureDevice)
	}
	return humitureDevices, nil
}

//GetDeviceHistory ..
func GetDeviceHistory(limit, offset int, devEUI string, startTime, endTime time.Time) ([]HumitureLog, error) {
	q := influxdb.Query{
		Command:  fmt.Sprintf(`SELECT * FROM %s WHERE ("DevEUI" = '%s') AND time >= %ds and time <= %ds limit %d offset %d`, setting.Cfg.Influxdb.TableName, devEUI, startTime.Unix(), endTime.Unix(), limit, offset),
		Database: setting.Cfg.Influxdb.Database,
	}
	response, err := Conn.Query(q)
	if err == nil {
		if response.Error() != nil {
			return nil, err
		}
		// fmt.Println(*response)
	} else {
		return nil, err
	}
	humitureLogs := make([]HumitureLog, 0)
	if response.Results == nil {
		return nil, errors.New("not hava data")
	}
	if response.Results[0].Series == nil {
		return nil, errors.New("not hava data")
	}
	for _, row := range response.Results[0].Series[0].Values {
		t, _ := time.Parse(time.RFC3339, row[0].(string))
		t = t.UTC().Add(time.Hour * 8)
		humitureLog := HumitureLog{
			UpDate:      t,
			DevEUI:      row[1].(string),
			DevName:     row[2].(string),
			Temperature: row[5].(json.Number),
			Humidity:    row[4].(json.Number),
			Electricity: row[3].(json.Number),
		}
		humitureLogs = append(humitureLogs, humitureLog)
	}
	return humitureLogs, nil
}

//GetDeviceHistoryCount ..
func GetDeviceHistoryCount(devEUI string, startTime, endTime time.Time) interface{} {
	q := influxdb.Query{
		Command:  fmt.Sprintf(`SELECT count(time) FROM %s WHERE ("DevEUI" = '%s') AND time >= %ds and time <= %ds`, setting.Cfg.Influxdb.TableName, devEUI, startTime.Unix(), endTime.Unix()),
		Database: setting.Cfg.Influxdb.Database,
	}
	response, err := Conn.Query(q)
	if err == nil {
		if response.Error() != nil {
			return nil
		}
		// fmt.Println(*response)
	} else {
		return nil
	}
	if response.Results == nil {
		return nil
	}
	if response.Results[0].Series == nil {
		return nil
	}
	for _, row := range response.Results[0].Series[0].Values {
		return row[2]
	}
	return nil
}

// GetDeviceChart ..
func GetDeviceChart(limit, offset int, devEUI string) ([]HumitureLog, error) {
	q := influxdb.Query{
		Command:  fmt.Sprintf(`SELECT * FROM %s WHERE ("DevEUI" = '%s') order by time desc limit %d offset %d`, setting.Cfg.Influxdb.TableName, devEUI, limit, offset),
		Database: setting.Cfg.Influxdb.Database,
	}
	response, err := Conn.Query(q)
	if err == nil {
		if response.Error() != nil {
			return nil, err
		}
		// fmt.Println(*response)
	} else {
		return nil, err
	}
	humitureLogs := make([]HumitureLog, 0)
	if response.Results == nil {
		return nil, errors.New("not hava data")
	}
	if response.Results[0].Series == nil {
		return nil, errors.New("not hava data")
	}
	for _, row := range response.Results[0].Series[0].Values {
		t, _ := time.Parse(time.RFC3339, row[0].(string))
		t = t.UTC().Add(time.Hour * 8)
		humitureLog := HumitureLog{
			UpDate:      t,
			DevEUI:      row[1].(string),
			DevName:     row[2].(string),
			Temperature: row[5].(json.Number),
			Humidity:    row[4].(json.Number),
			Electricity: row[3].(json.Number),
		}
		humitureLogs = append(humitureLogs, humitureLog)
	}
	return humitureLogs, nil
}
