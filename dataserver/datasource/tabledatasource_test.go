package datasource

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
	//_ "github.com/mattn/go-oci8"
	"os"
	"testing"
)

var ids *TableDataSource = nil

func TestMain(m *testing.M) {
	fmt.Println("Before ====================")
	err := orm.RegisterDataBase("default", "mysql", "tong:123456@tcp(127.0.0.1:3306)/idb", 30)
	err = orm.RegisterDataBase("pest", "mysql", "tong:123456@tcp(127.0.0.1:3306)/pest", 30)
	DBAlias2DBTypeContainer["default"] = "mysql"
	DBAlias2DBTypeContainer["pest"] = "mysql"
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	code := m.Run()
	fmt.Println("End ====================")
	os.Exit(code)
}

func createIDS() *TableDataSource {
	if ids == nil {
		ids = &TableDataSource{
			DBDataSource: DBDataSource{
				DataSource: DataSource{
					Name: "JEDA_USER",
					Field: []*MyProperty{
						&MyProperty{
							Name: "ORG_ID",
						},
						&MyProperty{
							Name:    "ORG_NAME",
							OutJoin: true,
							OutJoinDefine: &OutFieldProperty{
								Source:     createIDS_ORG(),
								ValueField: "ORG_NAME",
								JoinField:  "ORG_ID",
								ValueFunc:  nil,
							},
						},
						&MyProperty{
							Name: "USER_ID",
						},
					},
				},
				DBAlias: "default",
			},
			TableName: "JEDA_USER",
		}
		ids.Init()
	}
	return ids
}
func createIDS_ORG() *TableDataSource {
	idso := &TableDataSource{
		DBDataSource: DBDataSource{
			DataSource: DataSource{
				Name: "JEDA_ORG",
				Field: []*MyProperty{
					&MyProperty{
						Name: "ORG_ID",
					},
					&MyProperty{
						Name: "ORG_NAME",
					},
				},
			},
			DBAlias: "default",
		},
		TableName: "JEDA_ORG",
	}
	idso.Init()
	return idso
}

func createIDS_river() *TableDataSource {
	if ids == nil {
		ids = &TableDataSource{
			DBDataSource: DBDataSource{
				DataSource: DataSource{
					Name: "default",
				},
				DBAlias: "default",
			},
			TableName: "ST_RIVER_R",
		}
		ids.Init()
	}
	return ids
}

func printRS(rs *DataResultSet) {
	fmt.Println("=======================================================================")
	for k, v := range rs.Fields {
		fmt.Printf("%s:%s:%s\n", k, v.FieldType, strconv.Itoa(v.Index))
	}
	fmt.Println("=======================================================================")
	for _, row := range rs.Data {
		for _, item := range row {
			fmt.Printf("%s\t", item)
		}
		fmt.Println()
	}
}

func TestRiverIDS(t *testing.T) {
	ids := createIDS_river()
	ids.RowsLimit = 1000
	ids.AddCriteria("Z", OPER_LT, 5.00)
	ids.AddAggre("CNT", &AggreType{
		Predicate: AGG_COUNT,
		ColName:   "Z",
	})
	rs, _ := ids.DoFilter()
	printRS(rs)
}

func TestTableDataSourceGetAllData(t *testing.T) {
	ids := createIDS()
	rs, _ := ids.GetAllData()
	printRS(rs)
	ids.RowsLimit = 10
	rs, _ = ids.GetAllData()

	fmt.Println("===============================================================================================")
	printRS(rs)

	data, err := json.Marshal(ids)
	fmt.Println(err)
	fmt.Println(string(data))
}

func TestAddCriteria(t *testing.T) {
	ids := createIDS()
	//	var inf interface{}
	//	inf = ids
	ids.AddCriteria("ORG_ID", OPER_EQ, "001031")
	rs, err := ids.DoFilter()
	if err != nil {
		fmt.Print(err)
	}
	printRS(rs)
	//fmt.Println(reflect.TypeOf(inf))
	//fmt.Println(reflect.ValueOf(inf))
	//_,e:=inf.(TableDataSource)
	//fmt.Println(e)
}

//func TestSQLDataSource_GetAllData(t *testing.T) {
//	var ids *SQLDataSource
//	ids = &SQLDataSource{
//		DBDataSource: DBDataSource{
//			DataSource: DataSource{
//				Name: "JEDAUSERSLE",
//			},
//			DBAlias:        "default",
//			AutoFillFields: true,
//		},
//		ParamsValues: []interface{}{"001031"},
//		SQL:          "select * from  JEDA_USER where org_id=?",
//	}
//	ids.Init()
//	ids.AddCriteria("USER_ID", OPER_EQ, "ceshiwpj")
//	rs, _ := ids.DoFilter()
//	for _, item := range ids.Field {
//		fmt.Printf("%s\t%s\n", item.Name, item.DataType)
//	}
//	printRS(rs)
//}