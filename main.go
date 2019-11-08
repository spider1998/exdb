package main

import (
	"flag"
	"log"

	"exdb/db"
)

var (
	DSN string
)

func main() {
	var (
		input db.Input
		err   error
	)
	flag.StringVar(&input.User, "u", "", "user")
	flag.StringVar(&input.Pwd, "p", "", "password")
	flag.StringVar(&input.DataBase, "d", "", "database")
	flag.StringVar(&input.IP, "ip", "", "ip")
	flag.StringVar(&input.Port, "port", "", "port")
	flag.StringVar(&input.Table, "t", "", "table")
	flag.StringVar(&input.SQL, "sql", "", "sql")
	flag.IntVar(&input.Row, "row", 0, "row")
	flag.IntVar(&input.Col, "col", 0, "col")
	flag.StringVar(&input.File, "f", "", "file")
	flag.StringVar(&input.Designation, "des", "", "des")
	flag.StringVar(&input.Default, "df", "", "df")
	flag.BoolVar(&input.Import, "i", false, "import")
	flag.BoolVar(&input.Export, "e", false, "export")
	flag.Parse()
	input.CheckInput()
	input.DB, err = db.OpenDB(input.DSN)
	if err != nil {
		log.Fatal("Unable to connect to the database: " + err.Error())
		return
	}
	log.Println("Successfully connected to the database :" + input.DSN)
	if !input.Import && input.Export { //导出
		err = input.ExportDExcel()
		if err != nil {
			log.Fatal("Export data file failed: " + err.Error())
			return
		}
	}
	if input.Import && !input.Export { //导入
		err = input.ImportDExcel()
		if err != nil {
			log.Fatal("Import data file failed: " + err.Error())
			return
		}
	}
	return
}
