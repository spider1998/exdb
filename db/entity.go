package db

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/rs/xid"
	"github.com/shakinm/xlsReader/xls"
)

const (
	_xlsSuffix  = "xls"
	_xlsxSuffix = "xlsx"
	_timeFormat = "2006-01-02 15:04:05"
)

type Input struct {
	DSN         string       `json:"dsn"`             //地址
	SQL         string       `json:"sql"`             //导出条件语句（默认全部）
	File        string       `json:"file"`            //输出路径
	Row         int          `json:"row" default:"0"` //导入起始行
	Col         int          `json:"col" default:"0"` //导入起始列
	Import      bool         `json:"import"`          //导入
	Designation string       `json:"designation"`     //导入字段
	Default     string       `json:"default"`         //指定默认字段
	Export      bool         `json:"output"`          //导出
	User        string       `json:"user"`            //用户
	Pwd         string       `json:"pwd"`             //密码
	IP          string       `json:"ip"`              //ip
	Port        string       `json:"port"`            //端口
	DataBase    string       `json:"data_base"`       //数据库
	Table       string       `json:"table"`           //数据表
	DB          *xorm.Engine `json:"db"`
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
}

func (i *Input) ImportDExcel() (err error) {
	var (
		cols          string
		values        string
		defaultAppVal []string
		count         int
		res           [][]string
	)
	if i.Designation == "" {
		log.Fatal("Should designation import Filed In order,Specified by -des,example(name,age,year)")
		return
	} else {
		i.Designation = strings.Replace(i.Designation, " ", ",", -1)
		cols = strings.Join(strings.Split(i.Designation, ","), ",")
		cols = "(" + cols

	}
	cnt, err := i.DB.Table(i.Table).Count()
	if err != nil {
		return
	}
	count = int(cnt) + 1
	if i.File == "" {
		log.Fatal("Should specify import file,Specified by -f")
		return
	}
	f, err := os.Open(i.File)
	if err != nil {
		log.Fatal("file open error :" + err.Error())
		return
	}
	defer f.Close()
	if i.Default == "" {
		log.Fatal("Should designation some import Filed default value,Specified by -df,example(id:int,create_time:now,state:1)")
		return
	} else {
		i.Default = strings.Replace(i.Default, " ", ",", -1)
		i.Default = strings.Replace(i.Default, "：", ":", -1)
		ar := strings.Split(i.Default, ",")
		for k, a := range ar {
			ks := strings.Split(a, ":")
			if k == len(ar)-1 {
				cols += "," + ks[0] + ")"
			} else {
				cols += "," + ks[0]
			}
			defaultAppVal = append(defaultAppVal, ks[1])
		}
	}

	fname := strings.Split(i.File, ".")
	if fname[len(fname)-1] == _xlsSuffix {
		res, err = importXls(f, i.Row, i.Col)
		if err != nil {
			log.Fatal("file format incorrect")
			return
		}
	} else if fname[len(fname)-1] == _xlsxSuffix {
		res, err = importXlsx(f, i.Row, i.Col)
		if err != nil {
			log.Fatal("file format incorrect")
			return
		}
	} else {
		log.Fatal("file format incorrect +")
		return
	}
	for i, re := range res {
		var val string
		for j, r := range re {
			if j == 0 {
				val = "(" + "'" + r + "'"
			} else if j == len(re)-1 {
				val += "," + "'" + r + "'" + ","
				//追加默认字段值
				for di, deval := range defaultAppVal {
					var s = ","
					if di == len(defaultAppVal)-1 {
						s = ""
					}
					if deval == "int" {
						val += "'" + strconv.Itoa(count) + "'" + s
					} else if deval == "now" {
						val += "'" + time.Now().Format(_timeFormat) + "'" + s
					} else if deval == "string" {
						val += "'" + xid.New().String() + "'" + s
					} else {
						val += "'" + deval + "'" + s
					}
				}
				if i == len(res)-1 {
					val += ")"
				} else {
					val += "),"
				}
			} else {
				val += "," + "'"+r + "'" + ","
			}
		}
		values += val
	}
	sql := "insert into " + i.Table + cols + " values" + values
	_, err = i.DB.Exec(sql)
	if err != nil {
		log.Fatal("file to insert data :" + err.Error())
		return
	}
	return

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
	if i.File != "" {
		outPut = i.File
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

func importXls(f multipart.File, row int, col int) (res [][]string, err error) {
	ff, _ := os.Create("test.xls")
	_, err = io.Copy(ff, f)
	if err != nil {
		return
	}
	defer os.Remove("test.xls")
	workbook, err := xls.OpenFile("test.xls")
	if err != nil {
		log.Panic(err.Error())
	}
	sheet, err := workbook.GetSheet(0)
	if err != nil {
		log.Panic(err.Error())
	}
	var result [][]string
	for i := row; i <= sheet.GetNumberRows(); i++ {
		var re []string
		if row, err := sheet.GetRow(i); err == nil {
			for _, col := range row.GetCols() {
				re = append(re, col.GetString())
			}
			result = append(result, re)
		}
	}
	for _, re := range result {
		res = append(res, re[col:])
	}
	return
}

func importXlsx(f multipart.File, row int, col int) (res [][]string, err error) {
	var (
		sheet = "Sheet1"
	)
	xlsx, err := excelize.OpenReader(f)
	if err != nil {
		return
	}
	result := xlsx.GetRows(sheet)[row:]
	for _, re := range result {
		res = append(res, re[col:])
	}
	return
}
