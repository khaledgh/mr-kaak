package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

// mysqlErrDupEntry is MySQL's error number for a duplicate unique/primary key.
const mysqlErrDupEntry = 1062

// IsDuplicateKey reports whether err is a MySQL duplicate-key violation. Lets
// services translate a race-condition insert into a friendly conflict error.
func IsDuplicateKey(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == mysqlErrDupEntry
	}
	return false
}
