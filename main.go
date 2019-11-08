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
		help  bool
	)
	flag.BoolVar(&help, "h", false, "help")
	flag.StringVar(&input.User, "u", "", "database username")
	flag.StringVar(&input.Pwd, "p", "", "database password")
	flag.StringVar(&input.DataBase, "d", "", "the database to be connected")
	flag.StringVar(&input.IP, "ip", "", "database ip address")
	flag.StringVar(&input.Port, "port", "", "database port")
	flag.StringVar(&input.Table, "t", "", "the table to be connected")
	flag.StringVar(&input.SQL, "sql", "", "additional query conditions are required when exporting（sql）")
	flag.IntVar(&input.Row, "row", 0, "the starting row of the form official data when importing")
	flag.IntVar(&input.Col, "col", 0, "the starting col of the form official data when importing")
	flag.StringVar(&input.File, "f", "", "file path/file name when importing or exporting")
	flag.StringVar(&input.Designation, "des", "", "database fields corresponding to the table at import time (in order)")
	flag.StringVar(&input.Default, "df", "", "default fields to be set when importing (eg id, create_time, state)")
	flag.BoolVar(&input.Import, "i", false, "import data")
	flag.BoolVar(&input.Export, "e", false, "export data")
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
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
