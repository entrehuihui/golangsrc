package models

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// var schema = `create  table if not exists humiture_log(
// 	id bigserial primary key,
// 	dev_eui varchar(50) not null,
// 	dev_name varchar(100) not null,
// 	temperature float, -- 温度
// 	humidity float,  -- 湿度
// 	electricity float, --
// 	up_date timestamp, -- 上传时间
// 	created_at timestamp
// 	);`

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

type HumitureDevice struct {
	DevEUI  string `db:"dev_eui" json:"dev_eui"`
	DevName string `db:"dev_name" json:"dev_name"`
}

// func CreateTable(db *sqlx.DB) {
// 	db.MustExec(schema)
// }

func (item HumitureLog) Validate() error {
	return nil
}

// new log
func InsertHumiture(db sqlx.Ext, item *HumitureLog) error {
	if err := item.Validate(); err != nil {
		return err
	}

	_, err := db.Exec(`
		insert into  humiture_log
        (dev_eui,dev_name,temperature,humidity,electricity,up_date,created_at)
		values($1,$2,$3,$4,$5,$6,$7);`,
		item.DevEUI,
		item.DevName,
		item.Temperature,
		item.Humidity,
		item.Electricity,
		item.UpDate,
		item.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func GetHumitures(db sqlx.Queryer, limit, offset int, devEUI string) ([]HumitureLog, error) {
	var humitureLogs []HumitureLog

	err := sqlx.Select(db, &humitureLogs, `
			select * from ( 
			select id,dev_eui,dev_name,temperature,humidity,electricity,up_date
			from humiture_log 
			where dev_eui=$1
			order by id desc
			limit $2 offset $3
			) as t 
			order by id`, devEUI, limit, offset)
	if err != nil {
		return nil, err
	}

	return humitureLogs, nil
}

func GetHumituresHistory(db sqlx.Queryer, limit, offset int, devEUI string, start, end time.Time) ([]HumitureLog, error) {
	var humitureLogs []HumitureLog
	err := sqlx.Select(db, &humitureLogs, `
			select id,dev_eui,dev_name,temperature,humidity,electricity,up_date
			from humiture_log 
			where dev_eui=$1
			and up_date between $4 and $5
			order by up_date desc
			limit $2 offset $3
			`, devEUI, limit, offset, start, end)
	if err != nil {
		return nil, err
	}

	// 转东八区时间
	for idx, _ := range humitureLogs {
		humitureLogs[idx].UpDate = humitureLogs[idx].UpDate.UTC().Add(time.Hour * 8)
	}
	return humitureLogs, nil
}

func GetHumituresHistoryCount(db sqlx.Queryer, devEUI string, start, end time.Time) (int32, error) {
	var count int32

	err := sqlx.Get(db, &count, `
			select count(id) cnt
			from humiture_log 
			where dev_eui=$1
			and up_date between $2 and $3
			`, devEUI, start, end)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetDevices(db sqlx.Queryer) ([]HumitureDevice, error) {
	var res []HumitureDevice
	err := sqlx.Select(db, &res, `
		select a.dev_eui,a.dev_name from humiture_log a
		inner join (select dev_eui,max(id) maxid from humiture_log
					group by dev_eui) b
		on a.id=b.maxid`)
	if err != nil {
		return nil, err
	}
	return res, nil
}
