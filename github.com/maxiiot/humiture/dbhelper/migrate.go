package dbhelper

import (
	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate"
)

var schema = `create  table if not exists humiture_log(
	id bigserial primary key,
	dev_eui varchar(50) not null,
	dev_name varchar(100) not null,
	temperature float, -- 温度
	humidity float,  -- 湿度
	electricity float, --
	up_date timestamp with time zone, -- 上传时间
	created_at timestamp with time zone
	);`

// table init
func MigrateHumiture(db *sqlx.DB) error {
	var migrations = &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			&migrate.Migration{
				Id:   "123_humiture_log",
				Up:   []string{schema},
				Down: []string{"drop table humiture_log"},
			},
		},
	}
	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
