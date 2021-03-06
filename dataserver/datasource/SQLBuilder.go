package datasource

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var mu sync.Mutex

const (
	// OperEq 等于
	OperEq string = "="
	// OperNoteq 不等于
	OperNoteq string = "<>"
	// OperGt 大于
	OperGt string = ">"
	// OperLt 小于
	OperLt string = "<"
	// OperGtEg 大于等于
	OperGtEg string = ">="
	// OperLtEg 小于等于
	OperLtEg string = "<="
	// OperBetween 介于--之间
	OperBetween string = "BETWEEN"
	// OperIn 包含
	OperIn          string = "in"
	OperIsNull      string = "is null"
	OperIsNotNull   string = "is not null"
	OperAlwaysFalse string = "alwaysfalse"
	OperAlwaysTrue  string = "alwaystrue"
)

const (
	// CompAnd 与
	CompAnd string = "and"
	// CompOr 或
	CompOr string = "or"
	// CompNot 非
	CompNot string = "not"
	// CompNone 未知
	CompNone string = ""
)

// SQLCriteria SQL查询条件
type SQLCriteria struct {
	PropertyName string
	Operation    string
	Value        interface{}
	Complex      string
}

// AggreType 聚合类型
type AggreType struct {
	// Predicate 谓词
	Predicate int
	// ColName 字段名
	ColName string
}

const (
	// AggCount 计数
	AggCount int = 1
	// AggSum 求和
	AggSum int = 2
	// AggAvg 求算术平均
	AggAvg int = 3
	// AggMax 最大值
	AggMax int = 4
	// AggMin 最小值
	AggMin int = 5
)

const (
	INNER_JOIN = "INNER"
	INNER_LEFT = "LEFT"
)

type IAddCriteria interface {
	AddCriteria(field, operation, complex string, value interface{}) IAddCriteria
}

// 描述一个Join 字句
type PieceJoin struct {
	Join      string
	TableName string
	criteria  []*SQLCriteria
	OutField  []string
}

// ISQLBuilder SQL构造器接口
type ISQLBuilder interface {
	AddCriteria(field, operation, complex string, value interface{}) IAddCriteria
	AddJoin(jp *PieceJoin)
	CreateSelectSQL() (string, []interface{})
	CreateInsertSQLByMap(fieldvalues map[string]interface{}) (string, []interface{})
	CreateDeleteSQL() (string, []interface{})
	CreateUpdateSQL(fieldvalues map[string]interface{}) (string, []interface{})
	CreateKeyFieldsSQL() string
	CreateGetColsSQL() string
	ClearCriteria()
	AddAggre(outfield string, aggreType *AggreType)
}

// SQLBuilder SQL构造器类
type SQLBuilder struct {
	//表名
	tableName string
	//抽象表，ObjectTable为一个SQL语句，返回一个结果集，这个结果集作为查询的表参与Select语句，相当于select * from (ObjectTable) as tableName
	objectTable string
	//字段名
	columns []string
	//排序字段
	orderBy    []string
	criteria   []*SQLCriteria
	joinpiece  []*PieceJoin
	rowsLimit  int
	rowsOffset int
	aggre      map[string]*AggreType
}

// 条件中的特殊值，该类型的值表示引用SQL语句中其他表的字段
type FieldNameWithTableName struct {
	Tablename string
	Fielname  string
}

// MySQLSQLBuileder MySQl的SQL构造器
type MySQLSQLBuileder struct {
	SQLBuilder
}

// CreateSQLBuileder2ObjectTable 创建SQL构造器
func CreateSQLBuileder2ObjectTable(dbType string, objectTable string, tablename string, columns []string, orderby []string, rowslimit int, rowsoffset int) (ISQLBuilder, error) {
	switch dbType {
	case DbTypeMySQL:
		return &MySQLSQLBuileder{
			SQLBuilder: SQLBuilder{
				objectTable: objectTable,
				tableName:   tablename,
				columns:     columns,
				orderBy:     orderby,
				rowsLimit:   rowslimit,
				rowsOffset:  rowsoffset}}, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型" + dbType)
}

// CreateSQLBuileder2 创建SQL构造器
func CreateSQLBuileder2(dbType string, tablename string, columns []string, orderby []string, rowslimit int, rowsoffset int) (ISQLBuilder, error) {
	switch dbType {
	case DbTypeMySQL:
		return &MySQLSQLBuileder{
			SQLBuilder: SQLBuilder{
				tableName:  tablename,
				columns:    columns,
				orderBy:    orderby,
				rowsLimit:  rowslimit,
				rowsOffset: rowsoffset}}, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型" + dbType)
}

// CreateSQLBuileder 创建SQL构造器
func CreateSQLBuileder(dbType string, tablename string) (ISQLBuilder, error) {
	switch dbType {
	case DbTypeMySQL:
		return &MySQLSQLBuileder{
			SQLBuilder: SQLBuilder{
				tableName: tablename}}, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型" + dbType)
}

// AddCriteria
func (c *PieceJoin) AddCriteria(field, operation, complex string, value interface{}) IAddCriteria {
	c.criteria = append(c.criteria, &SQLCriteria{
		PropertyName: field,
		Operation:    operation,
		Value:        value,
		Complex:      complex,
	})
	return c
}

// AddJoin
func (c *MySQLSQLBuileder) AddJoin(jp *PieceJoin) {
	mu.Lock()
	defer mu.Unlock()
	if c.joinpiece == nil {
		c.joinpiece = make([]*PieceJoin, 0, 2)
	}
	c.joinpiece = append(c.joinpiece, jp)
}

// CreateKeyFieldsSQL 返回查询数据库表主键信息的SQL语句
func (c *MySQLSQLBuileder) CreateKeyFieldsSQL() string {
	if c.objectTable == "" {
		sqlstr := "SELECT a.column_name,b.data_type FROM INFORMATION_SCHEMA.`KEY_COLUMN_USAGE` a" +
			" inner join information_schema.columns b on a.table_name=b.table_name and a.column_name=b.column_name " +
			" WHERE a.table_name='" + c.tableName + "' AND a.constraint_name='PRIMARY'"
		return sqlstr
	}
	return ""
}

// CreateGetColsSQL 返回获取数据库表全部字段的SQL语句
func (c *MySQLSQLBuileder) CreateGetColsSQL() string {
	if c.objectTable == "" {
		return "SELECT column_name,data_type FROM information_schema.columns WHERE table_name='" + c.tableName + "'"
	}
	return ""
}

// ClearCriteria 清楚查询条件
func (c *MySQLSQLBuileder) ClearCriteria() {
	c.criteria = nil
}

// AddAggre 添加聚合
func (c *MySQLSQLBuileder) AddAggre(outfield string, aggreType *AggreType) {
	if c.aggre == nil {
		c.aggre = make(map[string]*AggreType)
	}
	c.aggre[outfield] = aggreType
}

// AddCriteria 删除条件
func (c *MySQLSQLBuileder) AddCriteria(field, operation, complex string, value interface{}) IAddCriteria {
	mu.Lock()
	defer mu.Unlock()
	if c.criteria == nil {
		c.criteria = make([]*SQLCriteria, 0, 10)
	}
	c.criteria = append(c.criteria, &SQLCriteria{
		PropertyName: field,
		Operation:    operation,
		Value:        value,
		Complex:      complex,
	})
	return c
}

// 生成条件子句
func (c *MySQLSQLBuileder) createCriteriaSubStr(tableName string, criteria []*SQLCriteria) (string, []interface{}) {
	var sqlwhere string
	param := make([]interface{}, 0, len(criteria))
	for i, cr := range criteria {
		fieldname := tableName + "." + cr.PropertyName
		if strings.Contains(cr.PropertyName, ".") {
			fieldname = cr.PropertyName
		}
		var exp string
		switch cr.Operation {
		case OperAlwaysFalse:
			exp = " 1=0 "
		case OperAlwaysTrue:
			exp = " 1=1 "
		case OperBetween:
			{
				switch reflect.TypeOf(cr.Value).Kind() {
				case reflect.Slice, reflect.Array:
					s := reflect.ValueOf(cr.Value)
					if s.Len() != 2 {
						panic("the BETWEEN operation in SQLBuilder the params must be array or slice, and length must be 2")
					}
					if f, ok := interface{}(s.Index(0).Interface()).(*FieldNameWithTableName); ok {
						exp = fmt.Sprint(fieldname, " BETWEEN "+f.Tablename+"."+f.Fielname+" and ")
					} else {
						exp = fmt.Sprint(fieldname, " BETWEEN ? and ")
						param = append(param, s.Index(0).Interface())
					}
					if f, ok := interface{}(s.Index(1).Interface()).(*FieldNameWithTableName); ok {
						exp = exp + f.Tablename + "." + f.Fielname
					} else {
						exp = exp + "?"
						param = append(param, s.Index(1).Interface())
					}
				default:
					{
						panic("the BETWEEN operation in SQLBuilder the params must be array or slice, and length must be 2")
					}
				}
			}
		case OperIn:
			{
				switch reflect.TypeOf(cr.Value).Kind() {
				case reflect.Slice, reflect.Array:
					s := reflect.ValueOf(cr.Value)
					ins := ""
					for si := 0; si < s.Len(); si++ {
						ins = ins + "?,"
						param = append(param, s.Index(si).Interface())
					}
					ins = strings.TrimRight(ins, ",")
					exp = fmt.Sprint(fieldname, " in (", ins, ")")
				default:
					{
						exp = fmt.Sprint(fieldname, " in (?)")
						param = append(param, cr.Value)
					}
				}
			}
		case OperIsNull:
			{
				exp = fmt.Sprint(fieldname, " is null ")
			}
		case OperIsNotNull:
			{
				exp = fmt.Sprint(fieldname, " is not null ")
			}
		default:
			{
				if f, ok := interface{}(cr.Value).(*FieldNameWithTableName); ok {
					exp = fmt.Sprint(fieldname, cr.Operation, f.Tablename, ".", f.Fielname)
				} else {
					exp = fmt.Sprint(fieldname, cr.Operation, "?")
					param = append(param, cr.Value)
				}
			}
		}
		if i != 0 {
			if cr.Complex == CompAnd || cr.Complex == CompOr {
				sqlwhere = fmt.Sprint(sqlwhere, " ", cr.Complex, " ", exp)
			}
		} else {
			sqlwhere = fmt.Sprint(sqlwhere, " ", exp)
		}
	}
	//sql += " WHERE " + sqlwhere
	return sqlwhere, param
}

// createWhereSubStr 创建查询Where语句
func (c *MySQLSQLBuileder) createWhereSubStr() (string, []interface{}) {
	sqlwhere, param := c.createCriteriaSubStr(c.tableName, c.criteria)
	return " WHERE " + sqlwhere, param
}

// CreateDeleteSQL 创建删除数据的SQL语句
func (c *MySQLSQLBuileder) CreateDeleteSQL() (string, []interface{}) {
	sql := "DELETE FROM " + c.tableName
	if c.criteria != nil {
		where, ps := c.createWhereSubStr()
		sql += where
		return sql, ps
	}
	return sql, nil
}

// CreateUpdateSQL 创建update语句
func (c *MySQLSQLBuileder) CreateUpdateSQL(fieldvalues map[string]interface{}) (string, []interface{}) {
	sql := "UPDATE " + c.tableName + " SET "
	params := make([]interface{}, len(fieldvalues), len(fieldvalues))
	i := 0
	for k, v := range fieldvalues {
		if i != 0 {
			sql += ","
		}
		sql += k + "=?"
		params[i] = v
		i++
	}
	if c.criteria != nil {
		where, ps := c.createWhereSubStr()
		sql += where
		params = append(params, ps...)
	}
	return sql, params
}

// CreateInsertSQLByMap 创建Insert语句
func (c *MySQLSQLBuileder) CreateInsertSQLByMap(fieldvalues map[string]interface{}) (string, []interface{}) {
	params := make([]interface{}, len(fieldvalues), len(fieldvalues))
	sql := "INSERT INTO " + c.tableName + " ("
	ps := ""
	i := 0
	for k, v := range fieldvalues {
		if i != 0 {
			sql += ","
			ps += ","
		}
		ps += "?"
		sql += k
		params[i] = v
		i++
	}
	sql = sql + ") VALUES (" + ps + ")"
	return sql, params
}

//处理链接
// inner join tablename on .......
func (c *MySQLSQLBuileder) createJoinSubStr() (string, []interface{}) {
	sql := ""
	ps := make([]interface{}, 0, 1)
	for _, pie := range c.joinpiece {
		if len(pie.criteria) == 0 {
			continue
		}
		if pie.Join == INNER_JOIN {
			sql += " inner join " + pie.TableName + " on "
		} else {
			sql += " left join " + pie.TableName + " on "
		}
		sqlwhere, param := c.createCriteriaSubStr(c.tableName, pie.criteria)
		sql += sqlwhere
		ps = append(ps, param)
	}
	return sql, ps
}

// CreateSelectSQL 创建Select语句
func (c *MySQLSQLBuileder) CreateSelectSQL() (string, []interface{}) {
	if c.objectTable != "" &&
		len(c.criteria) == 0 &&
		c.rowsLimit == 0 &&
		c.rowsOffset == 0 &&
		len(c.orderBy) == 0 &&
		len(c.aggre) == 0 &&
		(len(c.columns) == 0 || c.columns[0] == "*") {
		//符合上面条件的时候objectTable就是一条SQL语句，直接返回
		return c.objectTable, nil
	}
	var sql = "SELECT "
	var param []interface{}
	param = nil
	groupFields := make([]string, 0, 10)
	cols := c.columns
	if len(c.aggre) != 0 {
		//计算 group by子句中的字段列表
		if len(c.columns) != 0 {
			cols = make([]string, 0, 10)
			for _, col := range c.columns {
				if strings.Trim(col, " ") != "*" {
					cols = append(cols, c.tableName+"."+col)
					groupFields = append(groupFields, col)
				}
			}
		}
		//将聚合函数添加到选择字段列表
		for field, aggre := range c.aggre {
			var p string
			switch aggre.Predicate {
			case AggCount:
				p = "COUNT("
			case AggAvg:
				p = "AVG("
			case AggMax:
				p = "MAX("
			case AggMin:
				p = "MIN("
			case AggSum:
				p = "SUM("
			}
			p += c.tableName + "." + aggre.ColName + ") as " + field
			cols = append(cols, p)
		}
	}
	if len(cols) == 0 {
		//cols长度为0，选择*
		sql += c.tableName + ".* "
	} else {
		//生成选择的字段列表
		for i, fs := range cols {
			if i != 0 {
				sql += ","
			}
			sql += fs
		}
	}
	if len(c.joinpiece) != 0 {
		for _, jin := range c.joinpiece {
			if len(jin.OutField) == 0 {
				continue
			}
			for _, of := range jin.OutField {
				sql += "," + jin.TableName + "." + of
			}
		}
	}
	if c.objectTable == "" {
		sql += " FROM " + c.tableName
	} else {
		sql += " FROM (" + c.objectTable + ") as " + c.tableName
	}

	//处理链接
	// inner join tablename on .......
	if len(c.joinpiece) != 0 {
		insql, ps := c.createJoinSubStr()
		sql += insql
		param = append(param, ps)
	}

	if c.criteria != nil {
		where, ps := c.createWhereSubStr()
		sql += where
		param = append(param, ps...)
	}

	if len(groupFields) != 0 {
		var grs string
		for index, gr := range groupFields {
			if index != 0 {
				grs = fmt.Sprint(",", grs)
			}
			grs = fmt.Sprint(grs, c.tableName+"."+gr)
		}
		sql += " GROUP BY " + grs
	}

	if len(c.orderBy) != 0 {
		sql += " ORDER BY "
		for i, o := range c.orderBy {
			if i != 0 {
				sql += ","
			}
			sql += o
		}
	}

	if c.rowsLimit != 0 {
		sql += " LIMIT " + strconv.Itoa(c.rowsOffset) + "," + strconv.Itoa(c.rowsLimit)
	}

	return sql, param
}

////
////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////
