package common

import (
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

var SyncDevTime map[lorawan.EUI64]bool
