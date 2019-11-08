package db

import (
	"fmt"
	"log"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type Input struct {
	DSN      string       `json:"dsn"` //地址
	SQL      string       `json:"sql"`
	OutPut   string       `json:"out_put"`   //输出路径
	Import   bool         `json:"import"`    //导入
	Export   bool         `json:"output"`    //导出
	User     string       `json:"user"`      //用户
	Pwd      string       `json:"pwd"`       //密码
	IP       string       `json:"ip"`        //ip
	Port     string       `json:"port"`      //端口
	DataBase string       `json:"data_base"` //数据库
	Table    string       `json:"table"`     //数据表
	DB       *xorm.Engine `json:"db"`
}

func (i *Input) CheckInput() {
	if i.User == "" {
		log.Fatal("Database username empty,Specified by -u")
	}
	if i.Pwd == "" {
		log.Fatal("Database password empty,Specified by -p")
	}
	if i.IP == "" {
		log.Println("Database ip empty,Specified by -ip,default 127.0.0.1")
	}
	if i.Port == "" {
		log.Fatal("Database port empty,Specified by -port,default 3306")
	}
	if i.DataBase == "" {
		log.Fatal("Database empty,Specified by -d")
	}
	if i.Table == "" {
		log.Fatal("Database table empty,Specified by -t")
	}
	if i.Export && i.Import {
		log.Fatal(" Only specify one way: -i:import  -e:export")
	}
	if !i.Export && !i.Import {
		log.Fatal(" Should specify one way: -i:import  -e:export")
	}
	//拼接
	i.DSN = i.User + ":" + i.Pwd + "@tcp(" + i.IP + ":" + i.Port + ")/" + i.DataBase + "?charset=utf8mb4"
	if i.Import && !i.Export { //导入
		fmt.Println("===============")
	}
}

func (i *Input) ExportDExcel() (err error) {
	var (
		sql, outPut string
	)
	if i.SQL != "" {
		sql = i.SQL
	} else {
		sql = "select * from " + i.Table
	}
	result, err := i.DB.QueryString(sql)
	if err != nil {
		log.Fatal("db query error :【" + err.Error() + "】")
		return
	}
	f := excelize.NewFile()
	f.NewSheet("Sheet1")
	var (
		cnt  = 1
		cols []string
	)
	for k := range result[0] {
		f.SetCellValue("Sheet1", generateCell(cnt)+"1", k)
		cols = append(cols, k)
		cnt += 1
	}
	for j, vs := range result {
		var rows []interface{}
		for _, col := range cols {
			rows = append(rows, vs[col])
		}
		f.SetSheetRow("Sheet1", "A"+strconv.Itoa(j+2), &rows)
	}
	if i.OutPut != "" {
		outPut = i.OutPut
	} else {
		outPut = i.Table + ".xlsx"
	}
	err = f.SaveAs(outPut)
	if err != nil {
		log.Fatal("file save error :【" + err.Error() + "】")
		return
	}
	return
}

func generateCell(i int) (cell string) {
	switch i / 26 {
	case 0:
		cell = CellMap[i]
	case 1:
		cell = CellMap[i] + CellMap[26-i]
	default:
		panic("Data too long ...")
	}
	return
}
