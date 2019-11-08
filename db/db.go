package db

import (
	"os"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
)

func OpenDB(dsn string) (e *xorm.Engine, err error) {
	e, err = xorm.NewEngine("mysql", dsn)
	if err != nil {
		return
	}
	e.SetMapper(core.GonicMapper{})
	logger := xorm.NewSimpleLogger(os.Stdout)
	e.SetLogger(logger)
	e.ShowSQL(true)
	e.ShowExecTime(true)
	return
}