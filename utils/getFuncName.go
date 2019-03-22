package utils


import (
	"reflect"
	"runtime"
)


type DefineFunc func()


func getFuncName(fn DefineFunc) string {
	// 首先得到函数指针
	fp := reflect.ValueOf(fn).Pointer()

	// 获取函数名
	funcName := runtime.FuncForPC(fp).Name()

	// TODO
	// call func

	return funcName
}
