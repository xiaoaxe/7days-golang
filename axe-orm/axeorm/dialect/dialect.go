//dialect interface, 用于兼容不同的db实现，对应用层提供统一的访问接口
//@author: baoqiang
//@time: 2021/10/26 20:09:59
package dialect

import "reflect"

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string // 语言类型映射为db类型
	TableExistSQL(tableName string) (string, []interface{})
}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
