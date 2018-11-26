package main

import (
	"fmt"
	"log"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
)

// Conn ..
var Conn influxdb.Client

//Bp ..
var Bp influxdb.BatchPoints

func init() {
	conn, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     "http://127.0.0.1:8086",
		Username: "",
		Password: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	// for {
	// 	if _, _, err := Conn.Ping(time.Second * 3); err != nil {
	// 		fmt.Println("ping database error, will retry in 2s:", err)
	// 		time.Sleep(time.Second * 2)
	// 	} else {
	// 		break
	// 	}
	// }
	Conn = conn
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  "maxiiot",
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}
	Bp = bp
}

func main() {
	a := fmt.Sprintf(`SELECT * FROM %s WHERE ("DevEUI" = '%s') AND time >= %ds and time <= %ds limit %d offset %d`, "maxiiot", "1000000000000001", 1541650584, 1541670584, 10, 0)
	fmt.Println(a)
	q := influxdb.Query{
		Command:  a,
		Database: "maxiiot",
	}
	response, err := Conn.Query(q)
	if err == nil {
		if response.Error() != nil {
			return
		}
		// fmt.Println(*response)
	} else {
		return
	}
	humitureLogs := make([]HumitureLog, 0)
	for _, row := range response.Results[0].Series[0].Values {
		// fmt.Println(row)
		// fmt.Println(row[0])
		// t, err := time.Parse(time.RFC3339, row[0].(string))
		// if err != nil {
		// 	log.Fatal(err)
		// }
		t, _ := time.Parse(time.RFC3339, row[0].(string))
		t = t.UTC().Add(time.Hour * 8)
		valu := row[1]
		fmt.Println(t, valu, row[2], row[3], row[4])
		humitureLog := HumitureLog{
			DevEUI:  row[1].(string),
			DevName: row[2].(string),
		}
		humitureLogs = append(humitureLogs, humitureLog)
	}
	fmt.Println(humitureLogs)
}

type HumitureLog struct {
	ID          int64     `db:"id" json:"id"`
	DevEUI      string    `db:"dev_eui" json:"dev_eui"`
	DevName     string    `db:"dev_name" json:"dev_name"`
	Temperature float64   `db:"temperature" json:"temperature"`
	Humidity    float64   `db:"humidity" json:"humidity"`
	Electricity float64   `db:"electricity" json:"electricity"`
	UpDate      time.Time `db:"up_date" json:"up_date"`
	CreatedAt   time.Time `db:"created_at"`
}
